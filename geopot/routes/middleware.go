package routes

import (
	"net/http"
	"path"
	"strings"

	"jsfraz/geopot/utils"

	"github.com/gin-gonic/gin"
)

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

		// SPA fallback.
		c.Request.URL.Path = "/"
		c.Request.URL.RawPath = ""
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
