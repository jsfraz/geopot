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
}
