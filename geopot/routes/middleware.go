package routes

import (
	"compress/gzip"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"

	"jsfraz/geopot/utils"

	"github.com/gin-gonic/gin"
)

// Vite emits content-hashed asset names, which can be cached forever.
var hashedAssetPattern = regexp.MustCompile(`-[A-Za-z0-9_-]{8,}\.(?:js|css|woff2?|ttf|png|jpe?g|svg|webp)$`)

// Adds conservative security headers to every response.
//
//	@param debug
//	@return gin.HandlerFunc
func securityHeaders(debug bool) gin.HandlerFunc {
	scriptSrc := "'self'"
	styleSrc := "'self' 'unsafe-inline'"
	if debug {
		// Swagger UI is loaded from unpkg.com in debug mode only.
		scriptSrc += " https://unpkg.com"
		styleSrc += " https://unpkg.com"
	}

	contentSecurityPolicy := strings.Join([]string{
		"default-src 'self'",
		"script-src " + scriptSrc,
		"style-src " + styleSrc,
		"img-src 'self' data: blob: https://*.tile.openstreetmap.org https://tile.openstreetmap.org",
		"font-src 'self' data:",
		// WebSocket connections are same-origin but browsers match ws:/wss:
		// separately from http(s) for connect-src.
		"connect-src 'self' ws: wss:",
		"worker-src 'self' blob:",
		"child-src 'self' blob:",
		"object-src 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	}, "; ")

	return func(c *gin.Context) {
		headers := c.Writer.Header()
		headers.Set("Content-Security-Policy", contentSecurityPolicy)
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		headers.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		headers.Set("Cross-Origin-Opener-Policy", "same-origin")
		if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// Allows cross-origin reads only from explicitly allowed origins (or the same
// origin when no allowlist is configured). Credentials are never allowed.
//
//	@return gin.HandlerFunc
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && utils.IsAllowedOrigin(origin, c.Request.Host) {
			headers := c.Writer.Header()
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Add("Vary", "Origin")
			headers.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			headers.Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
			headers.Set("Access-Control-Max-Age", "43200")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Limits requests to the public API and WebSocket endpoint per client IP.
//
//	@param limiter
//	@return gin.HandlerFunc
func rateLimitMiddleware(limiter *utils.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if !strings.HasPrefix(requestPath, "/api") && requestPath != "/ws" {
			c.Next()
			return
		}

		if !limiter.Allow(c.ClientIP()) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// Compresses API responses. Static assets are left to the reverse proxy, which
// can cache the compressed representation and knows how to honour Range and
// conditional requests; the WebSocket upgrade needs the raw ResponseWriter so
// the connection can be hijacked.
//
// It must be registered after corsMiddleware so preflight requests, which are
// answered with an empty 204, never get a Content-Encoding header.
//
//	@return gin.HandlerFunc
func gzipMiddleware() gin.HandlerFunc {
	writers := sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(io.Discard)
		},
	}

	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") ||
			c.Request.Method == http.MethodHead ||
			!strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		writer := writers.Get().(*gzip.Writer)
		writer.Reset(c.Writer)
		defer func() {
			_ = writer.Close()
			writers.Put(writer)
		}()

		headers := c.Writer.Header()
		headers.Set("Content-Encoding", "gzip")
		headers.Add("Vary", "Accept-Encoding")
		headers.Del("Content-Length")

		c.Writer = &gzipResponseWriter{ResponseWriter: c.Writer, writer: writer}
		c.Next()
	}
}

// gzipResponseWriter writes the response body through a gzip.Writer.
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Writes the compressed body.
//
//	@receiver w
//	@param data
//	@return int
//	@return error
func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

// Writes the compressed body from a string.
//
//	@receiver w
//	@param s
//	@return int
//	@return error
func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.writer.Write([]byte(s))
}

// Returns the cache policy for a static file: content-hashed assets never
// change, everything else is revalidated daily.
//
//	@param requestPath
//	@return string
func cacheControlFor(requestPath string) string {
	if hashedAssetPattern.MatchString(requestPath) {
		return "public, max-age=31536000, immutable"
	}
	return "public, max-age=86400"
}

// Serves the built frontend from disk, with an SPA fallback to index.html.
// Directory listings are never exposed.
//
//	@param root
//	@return gin.HandlerFunc
func staticHandler(root string) gin.HandlerFunc {
	fileSystem := http.Dir(root)
	fileServer := http.FileServer(fileSystem)

	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		requestPath := c.Request.URL.Path
		// Unknown API/WebSocket paths must not fall back to the SPA.
		if strings.HasPrefix(requestPath, "/api") || requestPath == "/ws" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if file, err := fileSystem.Open(requestPath); err == nil {
			info, statErr := file.Stat()
			_ = file.Close()
			if statErr == nil {
				if !info.IsDir() {
					c.Header("Cache-Control", cacheControlFor(requestPath))
					fileServer.ServeHTTP(c.Writer, c.Request)
					return
				}
				// Never list directories.
				if requestPath != "/" {
					c.AbortWithStatus(http.StatusNotFound)
					return
				}
			}
		}

		// Do not return index.html for missing assets.
		if strings.HasPrefix(requestPath, "/assets/") || path.Ext(requestPath) != "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// SPA fallback: the shell must be revalidated on every load so a new
		// deployment is picked up immediately.
		c.Header("Cache-Control", "no-cache")
		c.Request.URL.Path = "/"
		c.Request.URL.RawPath = ""
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
