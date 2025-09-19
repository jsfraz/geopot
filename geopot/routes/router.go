package routes

import (
	"fmt"
	"jsfraz/geopot/handlers"
	"jsfraz/geopot/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

// Returns a new API router.
//
//	@return *fizz.Fizz
//	@return error
func NewRouter() (*fizz.Fizz, error) {
	// Gin instance
	engine := gin.New()
	// Logger middleware
	if utils.GetSingleton().Config.GinMode != "release" {
		engine.Use(gin.Logger())
	} else {
		// No Gin logging in release mode
	}
	// Recovery middleware
	engine.Use(gin.Recovery())
	// Default cors config, Allow Origin
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AllowHeaders = append(config.AllowHeaders, "x-requested-with") // Vue.js sends this header for some reason
	engine.Use(cors.New(config))

	// Fizz instance
	fizz := fizz.NewFromEngine(engine)

	// Base API route
	grp := fizz.Group("api", "", "")

	// OpenAPI spec
	if utils.GetSingleton().Config.GinMode != "release" {
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

	// Static files (frontend) - musí být až po definici všech API tras
	engine.NoRoute(func(c *gin.Context) {
		// Zkus servírovat statický soubor
		c.File("./static" + c.Request.URL.Path)
	})

	return fizz, nil
}
