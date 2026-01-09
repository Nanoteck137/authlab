package apis

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/nanoteck137/authlab/core"
	"github.com/nanoteck137/authlab/service"
	"github.com/nanoteck137/pyrin"
)

// TODO(patrik):
//  - Callback: Render HTML pages with success, error, expired
//  - Support for multiple oidc providers
//  - Add provider to users in database
//  - Code Login "ABCD-1234"
//  - Rename "AuthSession" to "AuthRequest"

type GetMe struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`

	DisplayName   string  `json:"displayName"`
	QuickPlaylist *string `json:"quickPlaylist"`
}

type AuthInitiate struct {
	SessionId string `json:"sessionId"`
	AuthUrl   string `json:"authUrl"`
	ExpiresAt string `json:"expiresAt"`
}

type AuthLoginWithCode struct {
	Token string `json:"token"`
}

type AuthLoginWithCodeBody struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type GetAuthCode struct {
	Code *string `json:"code"`
}

func InstallAuthHandlers(app core.App, group pyrin.Group) {
	group.Register(
		pyrin.ApiHandler{
			Name:         "AuthInitiate",
			Method:       http.MethodPost,
			Path:         "/auth/initiate",
			ResponseType: AuthInitiate{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				res, err := authService.CreateNormalSession()
				if err != nil {
					return nil, err
				}

				return AuthInitiate{
					SessionId: res.SessionId,
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

				userId, err := authService.GetUserFromCode(ctx, body.Code)
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

				err = authService.CompleteSession(state, code)
				if err != nil {
					if errors.Is(err, service.ErrAuthServiceSessionExpired) {
						return errors.New("session expired")
					}

					return err
				}

				return nil
			},
		},

		pyrin.ApiHandler{
			Name:         "GetAuthCode",
			Path:         "/auth/code/:sessionId",
			Method:       http.MethodGet,
			ResponseType: GetAuthCode{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				sessionId := c.Param("sessionId")

				authService, err := app.AuthService()
				if err != nil {
					return nil, err
				}

				code, err := authService.GetAuthCode(sessionId)
				if err != nil {
					if errors.Is(err, service.ErrAuthServiceSessionNotFound) {
						// TODO(patrik): Better error
						return nil, errors.New("session not found")
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

				// displayName := user.Username
				// if user.DisplayName.Valid {
				// 	displayName = user.DisplayName.String
				// }
				//
				// var quickPlaylist *string
				// if user.QuickPlaylist.Valid {
				// 	quickPlaylist = &user.QuickPlaylist.String
				// }

				return GetMe{
					Id: user.Id,
					// Username:      user.Username,
					Role: user.Role,
					// DisplayName:   displayName,
					// QuickPlaylist: quickPlaylist,
				}, nil
			},
		},
	)
}
