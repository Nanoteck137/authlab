package apis

import (
	"net/http"

	"github.com/nanoteck137/authlab"
	"github.com/nanoteck137/authlab/core"
	"github.com/nanoteck137/pyrin"
)

type GetSystemInfo struct {
	Version string `json:"version"`
}

func InstallSystemHandlers(app core.App, group pyrin.Group) {
	group.Register(
		pyrin.ApiHandler{
			Name:         "GetSystemInfo",
			Path:         "/system/info",
			Method:       http.MethodGet,
			ResponseType: GetSystemInfo{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				return GetSystemInfo{
					Version: authlab.Version,
				}, nil
			},
		},
	)
}
