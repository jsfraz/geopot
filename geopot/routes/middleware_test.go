package routes

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"jsfraz/geopot/utils"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func newTestRouter(t *testing.T, dir string) *gin.Engine {
	t.Helper()

	singleton := utils.GetSingleton()
	original := singleton.Config
	singleton.Config = utils.Config{GinMode: "release"}
	t.Cleanup(func() { singleton.Config = original })

	engine := gin.New()
	engine.Use(securityHeaders(false))
	engine.Use(corsMiddleware())
	engine.GET("/api/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	engine.NoRoute(staticHandler(dir))
	return engine
}

func TestStaticHandler(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>app</html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log(1)"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app-DMMvBMRI.js"), []byte("console.log(2)"), 0o600); err != nil {
		t.Fatal(err)
	}

	engine := newTestRouter(t, dir)

	cases := []struct {
		path         string
		status       int
		body         string
		bodyAbsent   bool
		cacheControl string
	}{
		{"/", http.StatusOK, "<html>app</html>", false, "no-cache"},
		{"/dashboard/route", http.StatusOK, "<html>app</html>", false, "no-cache"},
		{"/assets/app.js", http.StatusOK, "console.log(1)", false, "public, max-age=86400"},
		{"/assets/app-DMMvBMRI.js", http.StatusOK, "console.log(2)", false, "public, max-age=31536000, immutable"},
		{"/assets/", http.StatusNotFound, "", true, ""},
		{"/missing.js", http.StatusNotFound, "", true, ""},
		{"/api/unknown", http.StatusNotFound, "", true, ""},
	}

	for _, tc := range cases {
		request := httptest.NewRequest(http.MethodGet, tc.path, nil)
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, request)

		if recorder.Code != tc.status {
			t.Errorf("%s: got status %d, want %d", tc.path, recorder.Code, tc.status)
		}
		if tc.body != "" && recorder.Body.String() != tc.body {
			t.Errorf("%s: unexpected body %q", tc.path, recorder.Body.String())
		}
		if tc.bodyAbsent && recorder.Body.Len() != 0 {
			t.Errorf("%s: expected empty body, got %q", tc.path, recorder.Body.String())
		}
		if tc.cacheControl != "" {
			if got := recorder.Header().Get("Cache-Control"); got != tc.cacheControl {
				t.Errorf("%s: Cache-Control = %q, want %q", tc.path, got, tc.cacheControl)
			}
		}
	}
}

func TestCacheControlFor(t *testing.T) {
	cases := map[string]string{
		// Vite content-hashed assets never change.
		"/assets/index-DMMvBMRI.js":  "public, max-age=31536000, immutable",
		"/assets/index-CkmnwywG.css": "public, max-age=31536000, immutable",
		// Files copied verbatim from public/ keep their names.
		"/assets/earth-dark.jpg":                    "public, max-age=86400",
		"/assets/ne_110m_admin_0_countries.geojson": "public, max-age=86400",
		"/favicon.ico":                              "public, max-age=86400",
		"/fonts/FiraCodeNerdFontMono-Regular.ttf":   "public, max-age=86400",
	}

	for path, expected := range cases {
		if got := cacheControlFor(path); got != expected {
			t.Errorf("%s: got %q, want %q", path, got, expected)
		}
	}
}

func TestGzipMiddleware(t *testing.T) {
	engine := gin.New()
	engine.Use(corsMiddleware())
	engine.Use(gzipMiddleware())
	engine.GET("/api/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	engine.GET("/static.txt", func(c *gin.Context) { c.String(http.StatusOK, "plain") })

	// Compressed API response.
	request := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	reader, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatalf("response is not valid gzip: %v", err)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("body = %q", body)
	}

	// Static files are left to the reverse proxy.
	request = httptest.NewRequest(http.MethodGet, "/static.txt", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	recorder = httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Content-Encoding"); got != "" {
		t.Errorf("static file must not be compressed by the API middleware: %q", got)
	}

	// Preflight requests stay uncompressed.
	request = httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	recorder = httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Content-Encoding"); got != "" {
		t.Errorf("preflight response must not carry Content-Encoding: %q", got)
	}
}

func TestSecurityHeadersAndCORS(t *testing.T) {
	singleton := utils.GetSingleton()
	original := singleton.Config
	singleton.Config = utils.Config{GinMode: "release", AllowedOrigins: []string{"https://allowed.example"}}
	t.Cleanup(func() { singleton.Config = original })

	engine := gin.New()
	engine.Use(securityHeaders(false))
	engine.Use(corsMiddleware())
	engine.GET("/api/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// Same-origin request: headers present, no CORS header needed.
	request := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	request.Host = "geopot.example"
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("missing nosniff header: %q", got)
	}
	if got := recorder.Header().Get("Content-Security-Policy"); got == "" {
		t.Error("missing Content-Security-Policy header")
	}

	// Allowed cross-origin request.
	request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	request.Host = "geopot.example"
	request.Header.Set("Origin", "https://allowed.example")
	recorder = httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.example" {
		t.Errorf("allowed origin not echoed: %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("credentials must not be allowed: %q", got)
	}

	// Disallowed cross-origin request.
	request = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	request.Host = "geopot.example"
	request.Header.Set("Origin", "https://evil.example")
	recorder = httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("disallowed origin was echoed: %q", got)
	}
}
