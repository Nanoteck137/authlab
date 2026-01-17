package apis

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/maruel/natural"
	"github.com/nanoteck137/authlab/core"
	"github.com/nanoteck137/authlab/render"
	"github.com/nanoteck137/authlab/service"
	"github.com/nanoteck137/pyrin"
)

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

type AuthInitiateQuick struct {
	Code      string `json:"code"`
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

type AuthLoginQuickCodeBody struct {
	QuickCode string `json:"quickCode"`
}

type GetAuthTokenFromQuickCode struct {
	Token string `json:"token"`
}

type GetAuthTokenFromQuickCodeBody struct {
	Code string `json:"code"`
}

type AuthGetQuickCodeStatus struct {
	Status string `json:"status"`
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
			Name:         "AuthInitiateQuick",
			Method:       http.MethodPost,
			Path:         "/auth/initiate/quick",
			ResponseType: AuthInitiateQuick{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				res, err := authService.CreateQuickRequest()
				if err != nil {
					return nil, err
				}

				return AuthInitiateQuick{
					Code:    res.Code,
					AuthUrl: "FIX ME",
					// TODO(patrik): Move this to the AuthQuickRequest
					ExpiresAt: res.Expires.Format(time.RFC3339Nano),
				}, nil
			},
		},

		pyrin.ApiHandler{
			Name:     "AuthLoginQuickCode",
			Method:   http.MethodPost,
			Path:     "/auth/login-quick-code",
			BodyType: AuthLoginQuickCodeBody{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				body, err := pyrin.Body[AuthLoginQuickCodeBody](c)
				if err != nil {
					return nil, err
				}

				user, err := User(app, c)
				if err != nil {
					return nil, err
				}

				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				err = authService.CompleteQuickRequest(body.QuickCode, user.Id)
				if err != nil {
					return nil, err
				}

				return nil, nil
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

				err = authService.InvalidateRequest(body.State)
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
						render.RenderCallbackRequestExpired(c.Response())
						c.Response().WriteHeader(http.StatusOK)

						return nil
					}

					render.RenderCallbackError(c.Response())
					c.Response().WriteHeader(http.StatusOK)

					return nil
				}

				render.RenderCallbackSuccess(c.Response())
				c.Response().WriteHeader(http.StatusOK)

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
			Name:         "AuthGetQuickCodeStatus",
			Path:         "/auth/quick-code/status/:code",
			Method:       http.MethodGet,
			ResponseType: AuthGetQuickCodeStatus{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				code := c.Param("code")

				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				status, err := authService.CheckQuickRequestStatus(code)
				if err != nil {
					if errors.Is(err, service.ErrAuthServiceRequestNotFound) {
						// TODO(patrik): Better error
						return nil, errors.New("request not found")
					}

					return nil, err
				}

				return AuthGetQuickCodeStatus{
					Status: string(status),
				}, nil
			},
		},

		pyrin.ApiHandler{
			Name:         "AuthCreateQuickCodeToken",
			Path:         "/auth/quick-code/token",
			Method:       http.MethodPost,
			ResponseType: GetAuthTokenFromQuickCode{},
			BodyType:     GetAuthTokenFromQuickCodeBody{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				body, err := pyrin.Body[GetAuthTokenFromQuickCodeBody](c)
				if err != nil {
					return nil, err
				}

				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				token, err := authService.GetAuthTokenForQuickCode(body.Code)
				if err != nil {
					if errors.Is(err, service.ErrAuthServiceRequestNotFound) {
						// TODO(patrik): Better error
						return nil, errors.New("request not found")
					}

					return nil, err
				}

				return GetAuthTokenFromQuickCode{
					Token: token,
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
