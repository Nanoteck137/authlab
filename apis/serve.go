package apis

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nanoteck137/authlab"
	"github.com/nanoteck137/authlab/core"
	"github.com/nanoteck137/authlab/render"
	"github.com/nanoteck137/pyrin"
)

type hookedResponseWriter struct {
	http.ResponseWriter
	got404 bool
}

func (hrw *hookedResponseWriter) WriteHeader(status int) {
	if status == http.StatusNotFound {
		// Don't actually write the 404 header, just set a flag.
		hrw.got404 = true
	} else {
		hrw.ResponseWriter.WriteHeader(status)
	}
}

func (hrw *hookedResponseWriter) Write(p []byte) (int, error) {
	if hrw.got404 {
		// No-op, but pretend that we wrote len(p) bytes to the writer.
		return len(p), nil
	}

	return hrw.ResponseWriter.Write(p)
}

func RegisterHandlers(app core.App, router pyrin.Router) {
	g := router.Group("/api/v1")
	InstallAuthHandlers(app, g)
	InstallSystemHandlers(app, g)
	InstallUserHandlers(app, g)

	g = router.Group("")
	g.Register(
		pyrin.NormalHandler{
			Name:        "Test",
			Method:      http.MethodGet,
			Path:        "/test",
			HandlerFunc: func(c pyrin.Context) error {
				render.RenderCallbackError(c.Response())
				c.Response().WriteHeader(200)

				return nil
			},
		},

		pyrin.NormalHandler{
			Name:        "Test",
			Method:      http.MethodGet,
			Path:        "/static/*",
			HandlerFunc: func(c pyrin.Context) error {
				f := os.DirFS("./render/static")
				fs := http.StripPrefix("/static", http.FileServerFS(f))

				fs.ServeHTTP(c.Response(), c.Request())

				return nil
			},
		},

		pyrin.NormalHandler{
			Name:   "RootFiles",
			Method: http.MethodGet,
			Path:   "/*",
			HandlerFunc: func(c pyrin.Context) error {
				f := os.DirFS("./result")
				fs := http.FileServerFS(f)

				hookedWriter := &hookedResponseWriter{ResponseWriter: c.Response()}
				fs.ServeHTTP(hookedWriter, c.Request())

				if hookedWriter.got404 {
					if !strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
						c.Response().WriteHeader(http.StatusNotFound)
						fmt.Fprint(c.Response(), "404 not found")
					} else {
						c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
						pyrin.ServeFile(c, f, "index.html")
					}
				}

				return nil
			},
		},
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
