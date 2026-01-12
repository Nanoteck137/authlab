package apis

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/maruel/natural"
	"github.com/nanoteck137/authlab/core"
	"github.com/nanoteck137/authlab/service"
	"github.com/nanoteck137/pyrin"
)

// TODO(patrik):
//  - Callback: Render HTML pages with success, error, expired
//  - Support for multiple oidc providers
//  - Add provider to users in database
//  - Code Login "ABCD-1234"

type GetMe struct {
	Id          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type AuthInitiate struct {
	RequestId string `json:"requestId"`
	AuthUrl   string `json:"authUrl"`
	ExpiresAt string `json:"expiresAt"`
}

type AuthLoginWithCode struct {
	Token string `json:"token"`
}

type AuthLoginWithCodeBody struct {
	ProviderId string `json:"providerId"`
	Code       string `json:"code"`
	State      string `json:"state"`
}

type GetAuthCode struct {
	Code *string `json:"code"`
}

type AuthProvider struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type GetAuthProviders struct {
	Providers []AuthProvider `json:"providers"`
}

func InstallAuthHandlers(app core.App, group pyrin.Group) {
	group.Register(
		pyrin.ApiHandler{
			Name:         "GetAuthProviders",
			Method:       http.MethodGet,
			Path:         "/auth/providers",
			ResponseType: GetAuthProviders{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				providers := app.Config().OidcProviders

				res := GetAuthProviders{
					Providers: make([]AuthProvider, 0, len(providers)),
				}

				for id, provider := range providers {
					res.Providers = append(res.Providers, AuthProvider{
						Id:          id,
						DisplayName: provider.Name,
					})
				}

				sort.Slice(res.Providers, func(i, j int) bool {
					return natural.Less(res.Providers[i].DisplayName, res.Providers[j].DisplayName)
				})

				return res, nil
			},
		},

		pyrin.ApiHandler{
			Name:         "AuthInitiate",
			Method:       http.MethodPost,
			Path:         "/auth/initiate/:providerId",
			ResponseType: AuthInitiate{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				providerId := c.Param("providerId")

				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				res, err := authService.CreateNormalRequest(providerId)
				if err != nil {
					return nil, err
				}

				return AuthInitiate{
					RequestId: res.RequestId,
					AuthUrl:   res.AuthUrl,
					ExpiresAt: res.Expires.Format(time.RFC3339Nano),
				}, nil
			},
		},

		pyrin.ApiHandler{
			Name:         "AuthLoginWithCode",
			Method:       http.MethodPost,
			Path:         "/auth/login-with-code",
			ResponseType: AuthLoginWithCode{},
			BodyType:     AuthLoginWithCodeBody{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				body, err := pyrin.Body[AuthLoginWithCodeBody](c)
				if err != nil {
					return nil, err
				}

				ctx := context.TODO()

				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				userId, err := authService.GetUserFromCode(ctx, body.ProviderId, body.Code)
				if err != nil {
					return nil, err
				}

				token, err := authService.SignUserToken(userId)
				if err != nil {
					return nil, err
				}

				return AuthLoginWithCode{
					Token: token,
				}, nil
			},
		},

		pyrin.NormalHandler{
			Name:   "AuthCallback",
			Method: http.MethodGet,
			Path:   "/auth/callback",
			HandlerFunc: func(c pyrin.Context) error {
				url := c.Request().URL
				state := url.Query().Get("state")
				code := url.Query().Get("code")

				authService, err := app.AuthService()
				if err != nil {
					return err
				}

				err = authService.CompleteRequest(state, code)
				if err != nil {
					if errors.Is(err, service.ErrAuthServiceRequestExpired) {
						return errors.New("request expired")
					}

					return err
				}

				return nil
			},
		},

		pyrin.ApiHandler{
			Name:         "GetAuthCode",
			Path:         "/auth/code/:requestId",
			Method:       http.MethodGet,
			ResponseType: GetAuthCode{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				requestId := c.Param("requestId")

				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				code, err := authService.GetAuthCode(requestId)
				if err != nil {
					if errors.Is(err, service.ErrAuthServiceRequestNotFound) {
						// TODO(patrik): Better error
						return nil, errors.New("request not found")
					}

					return nil, err
				}

				return GetAuthCode{
					Code: code,
				}, nil
			},
		},

		pyrin.ApiHandler{
			Name:         "GetMe",
			Path:         "/auth/me",
			Method:       http.MethodGet,
			ResponseType: GetMe{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				user, err := User(app, c)
				if err != nil {
					return nil, err
				}

				return GetMe{
					Id:          user.Id,
					Email:       user.Email,
					DisplayName: user.DisplayName,
					Role:        user.Role,
				}, nil
			},
		},
	)
}
