package routes

import (
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

	engine := newTestRouter(t, dir)

	cases := []struct {
		path       string
		status     int
		body       string
		bodyAbsent bool
	}{
		{"/", http.StatusOK, "<html>app</html>", false},
		{"/dashboard/route", http.StatusOK, "<html>app</html>", false},
		{"/assets/app.js", http.StatusOK, "console.log(1)", false},
		{"/assets/", http.StatusNotFound, "", true},
		{"/missing.js", http.StatusNotFound, "", true},
		{"/api/unknown", http.StatusNotFound, "", true},
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
