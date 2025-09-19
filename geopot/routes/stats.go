package routes

import (
	"jsfraz/geopot/handlers"
	"jsfraz/geopot/utils"
	"net/http"

	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz"
)

// Sets stats route group.
//
//	@param grp
func StatsRoute(g *fizz.RouterGroup) {

	grp := g.Group("stats", "Stats", "Stats route")

	// Total connections
	grp.GET("totalConnections",
		utils.CreateOperationOption(
			"Total connection count",
			"Get total connection count",
			[]int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}),
		tonic.Handler(handlers.GetTotalConnectionCount, http.StatusOK),
	)

	// Total unique IPs
	grp.GET("totalUniqueIps",
		utils.CreateOperationOption(
			"Total unique IP count",
			"Get total unique IP count",
			[]int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}),
		tonic.Handler(handlers.GetTotalUniqueIPCount, http.StatusOK),
	)

	// All unique IPs
	grp.GET("allUniqueIps",
		utils.CreateOperationOption(
			"All unique IP addresses",
			"Get all unique IP addresses",
			[]int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}),
		tonic.Handler(handlers.GetAllUniqueIPAddresses, http.StatusOK),
	)

	// Total unique countries
	grp.GET("totalUniqueCountries",
		utils.CreateOperationOption(
			"Total unique countries count",
			"Get total unique countries count",
			[]int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}),
		tonic.Handler(handlers.GetTotalUniqueCountryCount, http.StatusOK),
	)

	// All unique countries
	grp.GET("allUniqueCountries",
		utils.CreateOperationOption(
			"All unique countries",
			"Get all unique countries",
			[]int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}),
		tonic.Handler(handlers.GetAllUniqueCountries, http.StatusOK),
	)
}
