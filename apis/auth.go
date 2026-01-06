package apis

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/kr/pretty"
	"github.com/nanoteck137/authlab/core"
	"github.com/nanoteck137/authlab/tools/utils"
	"github.com/nanoteck137/pyrin"
	"github.com/nanoteck137/validate"
	"golang.org/x/oauth2"
)

// TODO(patrik):
//  - Callback: Set the code on the session object
//  - Callback: Render HTML pages with success, error
//  - Create a poll mechanism for the frontend client to get the code

type Signup struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

// TODO(patrik): Test if this works with validation
type SignupBody struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}

var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9-]+$")
var passwordLengthRule = validate.Length(8, 32)

// TODO(patrik): Remove? and let the usernameRegex handle error
func (b *SignupBody) Transform() {
	b.Username = strings.TrimSpace(b.Username)
}

func (b SignupBody) Validate() error {
	checkPasswordMatch := validate.By(func(value interface{}) error {
		if b.PasswordConfirm != b.Password {
			return errors.New("password mismatch")
		}

		return nil
	})

	return validate.ValidateStruct(&b,
		validate.Field(&b.Username, validate.Required, validate.Length(4, 32), validate.Match(usernameRegex).Error("not valid username")),
		validate.Field(&b.Password, validate.Required, passwordLengthRule, checkPasswordMatch),
		validate.Field(&b.PasswordConfirm, validate.Required, checkPasswordMatch),
	)
}

type Signin struct {
	Token string `json:"token"`
}

type SigninBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (b SigninBody) Validate() error {
	return validate.ValidateStruct(&b,
		validate.Field(&b.Username, validate.Required),
		validate.Field(&b.Password, validate.Required),
	)
}

// TODO(patrik): Test if this works with validation
type ChangePasswordBody struct {
	CurrentPassword    string `json:"currentPassword"`
	NewPassword        string `json:"newPassword"`
	NewPasswordConfirm string `json:"newPasswordConfirm"`
}

func (b ChangePasswordBody) Validate() error {
	checkPasswordMatch := validate.By(func(value interface{}) error {
		if b.NewPasswordConfirm != b.NewPassword {
			return errors.New("password mismatch")
		}

		return nil
	})

	return validate.ValidateStruct(
		&b,
		validate.Field(&b.CurrentPassword, validate.Required),
		validate.Field(&b.NewPassword, validate.Required, passwordLengthRule, checkPasswordMatch),
		validate.Field(&b.NewPasswordConfirm, validate.Required, checkPasswordMatch),
	)
}

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
}

type AuthLoginWithCode struct {
	Token string `json:"token"`
}

type AuthLoginWithCodeBody struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type AuthSessionType string

const (
	AuthSessionTypeNormal    AuthSessionType = "normal"
	AuthSessionTypeQuickCode AuthSessionType = "quick-code"
)

type AuthSessionStatus string

const (
	AuthSessionStatusPending   AuthSessionStatus = "pending"
	AuthSessionStatusCompleted AuthSessionStatus = "completed"
	AuthSessionStatusExpired   AuthSessionStatus = "expired"
	AuthSessionStatusFailed    AuthSessionStatus = "failed"
)

type AuthSession struct {
	Id      string
	Type    AuthSessionType
	Status  AuthSessionStatus
	Expires time.Time
	Delete  time.Time
}

type AuthService struct {
	Sessions map[string]*AuthSession
}

func (a *AuthService) CreateNormalSession() *AuthSession {
	id := utils.CreateId()

	t := time.Now()
	session := &AuthSession{
		Id:      id,
		Type:    AuthSessionTypeNormal,
		Status:  AuthSessionStatusPending,
		Expires: t.Add(5 * time.Minute),
		Delete: t.Add(1 * time.Hour),
	}

	a.Sessions[id] = session

	return session
}

func (a *AuthService) RemoveUnusedEntries() {
	now := time.Now()
	for k, session := range a.Sessions {
		if session.Delete.After(now) {
			delete(a.Sessions, k)
		}
	}
}

func (a *AuthService) RunRoutine() {
    ticker := time.NewTicker(30 * time.Minute)
    for range ticker.C {
		a.RemoveUnusedEntries()
    }
}

var authService = &AuthService{
	Sessions: make(map[string]*AuthSession),
}

func InstallAuthHandlers(app core.App, group pyrin.Group) {
	if app != nil {
		go authService.RunRoutine()
	}

	group.Register(
		pyrin.ApiHandler{
			Name:         "AuthInitiate",
			Method:       http.MethodPost,
			Path:         "/auth/initiate",
			ResponseType: AuthInitiate{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				ctx := context.TODO()

				config := app.Config()

				provider, err := oidc.NewProvider(ctx, config.OdicIssuerUrl)
				if err != nil {
					return nil, fmt.Errorf("Failed to create OIDC provider: %w", err)
				}

				oauth2Config := &oauth2.Config{
					ClientID:     config.OdicClientId,
					ClientSecret: config.OdicClientSecret,
					RedirectURL:  config.OdicRedirectUrl,
					Endpoint:     provider.Endpoint(),
					Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
				}

				session := authService.CreateNormalSession()
				authURL := oauth2Config.AuthCodeURL(session.Id)

				return AuthInitiate{
					SessionId: session.Id,
					AuthUrl:   authURL,
				}, nil
			},
		},

		pyrin.ApiHandler{
			Name:         "AuthLoginWithCode",
			Method:       http.MethodPost,
			Path:         "/auth/loginWithCode",
			ResponseType: AuthLoginWithCode{},
			BodyType:     AuthLoginWithCodeBody{},
			HandlerFunc: func(c pyrin.Context) (any, error) {
				body, err := pyrin.Body[AuthLoginWithCodeBody](c)
				if err != nil {
					return nil, err
				}

				ctx := context.TODO()

				config := app.Config()

				provider, err := oidc.NewProvider(ctx, config.OdicIssuerUrl)
				if err != nil {
					return nil, fmt.Errorf("Failed to create OIDC provider: %w", err)
				}

				oauth2Config := &oauth2.Config{
					ClientID:     config.OdicClientId,
					ClientSecret: config.OdicClientSecret,
					RedirectURL:  config.OdicRedirectUrl,
					Endpoint:     provider.Endpoint(),
					Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
				}

				verifier := provider.Verifier(&oidc.Config{ClientID: config.OdicClientId})

				oauth2Token, err := oauth2Config.Exchange(ctx, body.Code)
				if err != nil {
					return nil, err
				}

				pretty.Println(oauth2Token)

				rawIDToken, ok := oauth2Token.Extra("id_token").(string)
				if !ok {
					return nil, errors.New("failed to login")
				}

				idToken, err := verifier.Verify(ctx, rawIDToken)
				if err != nil {
					return nil, err
				}

				// Extract user claims from OIDC token
				var oidcClaims struct {
					Email string `json:"email"`
					Name  string `json:"name"`
					Sub   string `json:"sub"`
				}
				if err := idToken.Claims(&oidcClaims); err != nil {
					return nil, err
				}

				pretty.Println(oidcClaims)

				// app.DB().GetUserById()

				// // Create YOUR OWN JWT token for the frontend to use
				// appToken, refreshToken, err := generateTokenPair(oidcClaims.Sub, oidcClaims.Email, oidcClaims.Name)
				// if err != nil {
				// }

				return AuthLoginWithCode{}, nil
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

				pretty.Println(url.Query())

				_ = code

				session := authService.Sessions[state]
				session.Status = AuthSessionStatusCompleted

				return nil
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

				displayName := user.Username
				if user.DisplayName.Valid {
					displayName = user.DisplayName.String
				}

				var quickPlaylist *string
				if user.QuickPlaylist.Valid {
					quickPlaylist = &user.QuickPlaylist.String
				}

				return GetMe{
					Id:            user.Id,
					Username:      user.Username,
					Role:          user.Role,
					DisplayName:   displayName,
					QuickPlaylist: quickPlaylist,
				}, nil
			},
		},
	)
}
