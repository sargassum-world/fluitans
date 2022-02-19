// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/controllers"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/dns"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/home"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/networks"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

var Pages []route.Templated = route.CollectTemplated(
	home.Pages,
	auth.Pages,
	controllers.Pages,
	networks.Pages,
	dns.Pages,
)
