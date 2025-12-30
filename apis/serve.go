package apis

import (
	"github.com/nanoteck137/authlab"
	"github.com/nanoteck137/authlab/core"
	"github.com/nanoteck137/pyrin"
)

func RegisterHandlers(app core.App, router pyrin.Router) {
	g := router.Group("/api/v1")
	InstallAuthHandlers(app, g)
	InstallSystemHandlers(app, g)
	InstallUserHandlers(app, g)
}

func Server(app core.App) (*pyrin.Server, error) {
	s := pyrin.NewServer(&pyrin.ServerConfig{
		LogName: authlab.AppName,
		RegisterHandlers: func(router pyrin.Router) {
			RegisterHandlers(app, router)
		},
	})

	return s, nil
}
