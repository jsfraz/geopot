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

	// Get the number of connections in the last 24 hours.
	grp.GET("last24HourConnections",
		utils.CreateOperationOption(
			"Last 24 Hour Connections",
			"Get the number of connections in the last 24 hours",
			[]int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}),
		tonic.Handler(handlers.GetLast24HourConnections, http.StatusOK),
	)

	// Get the server's own info.
	grp.GET("selfInfo",
		utils.CreateOperationOption(
			"Self Info",
			"Get the server's own info",
			[]int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}),
		tonic.Handler(handlers.GetServerInfo, http.StatusOK),
	)

	// Get all latitude and longitude pairs from the database.
	grp.GET("allLatLng",
		utils.CreateOperationOption(
			"All Latitude and Longitude",
			"Get all latitude and longitude pairs from the database",
			[]int{
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}),
		tonic.Handler(handlers.GetAllLatLng, http.StatusOK),
	)
}
