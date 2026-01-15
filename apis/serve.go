package apis

import (
	"net/http"
	"os"

	"github.com/nanoteck137/authlab"
	"github.com/nanoteck137/authlab/core"
	"github.com/nanoteck137/pyrin"
)

func RegisterHandlers(app core.App, router pyrin.Router) {
	g := router.Group("/api/v1")
	InstallAuthHandlers(app, g)
	InstallSystemHandlers(app, g)
	InstallUserHandlers(app, g)

	g = router.Group("")
	g.Register(
		pyrin.NormalHandler{
			Method:      http.MethodGet,
			Path:        "/static/*",
			HandlerFunc: func(c pyrin.Context) error {
				f := os.DirFS("./render/static")
				fs := http.StripPrefix("/static", http.FileServerFS(f))

				fs.ServeHTTP(c.Response(), c.Request())

				return nil
			},
		},

		pyrin.SpaHandler(os.DirFS("./result"), "index.html"),
	)
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
