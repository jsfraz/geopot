package routes

import (
	"errors"
	"fmt"
	"jsfraz/geopot/handlers"
	"jsfraz/geopot/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"

	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

// Returns a new API router.
//
//	@return *fizz.Fizz
//	@return error
func NewRouter() (*fizz.Fizz, error) {
	config := utils.GetSingleton().Config

	// Keep Gin's internal mode in sync with our config (release by default).
	if config.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Gin instance
	engine := gin.New()
	// Logger middleware (debug only, to avoid logging captured credentials in production)
	if config.IsDebug() {
		engine.Use(gin.Logger())
	}
	// Recovery middleware
	engine.Use(gin.Recovery())

	// Only trust forwarded headers from explicitly configured proxies.
	if err := engine.SetTrustedProxies(config.TrustedProxies); err != nil {
		return nil, fmt.Errorf("invalid TRUSTED_PROXIES: %w", err)
	}

	// Security headers, CORS and API rate limiting
	engine.Use(securityHeaders(config.IsDebug()))
	engine.Use(corsMiddleware())
	engine.Use(rateLimitMiddleware(utils.NewRateLimiter(config.APIRateLimitPerMinute)))

	// Never leak internal (e.g. database) errors to API clients.
	tonic.SetErrorHook(func(c *gin.Context, err error) (int, interface{}) {
		var bindError tonic.BindError
		if errors.As(err, &bindError) {
			return http.StatusBadRequest, gin.H{"error": "invalid request"}
		}
		log.Printf("API error on %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		return http.StatusInternalServerError, gin.H{"error": "internal server error"}
	})

	// Fizz instance
	fizz := fizz.NewFromEngine(engine)

	// Base API route
	grp := fizz.Group("api", "", "")

	// OpenAPI spec
	if config.IsDebug() {
		// Servers
		fizz.Generator().SetServers([]*openapi.Server{
			{
				Description: "localhost - debug",
				URL:         "http://localhost:8080",
			},
		})
		// TODO more info
		infos := &openapi.Info{
			Title:       "geopot",
			Description: "Monitoring SSH login attempts and geolocating remote hosts who failed to login and gathering used credentials.",
			Version:     "1.0.0",
			// TODO license
			Contact: &openapi.Contact{
				Name:  "Josef Ráž",
				URL:   "https://josefraz.cz",
				Email: "razj@josefraz.cz",
			},
			// TODO ToS
			// TODO XLogo
		}
		grp.GET("openapi.json", nil, fizz.OpenAPI(infos, "json"))
		// Swagger UI (https://github.com/swagger-api/swagger-ui/blob/HEAD/docs/usage/installation.md#unpkg)
		engine.LoadHTMLGlob("html/*.html")
		engine.GET("/swagger", func(c *gin.Context) {
			c.HTML(200, "swagger.html", nil)
		})
	}

	// WebSocket handler
	engine.GET("/ws", handlers.WebSocketHandler)

	// Setup other routes
	StatsRoute(grp)
	if len(fizz.Errors()) != 0 {
		return nil, fmt.Errorf("fizz errors: %v", fizz.Errors())
	}

	// Static files (frontend) - must be after defining all API routes
	engine.NoRoute(staticHandler("./static"))

	return fizz, nil
}
