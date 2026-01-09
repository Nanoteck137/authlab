package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kr/pretty"
	"github.com/nanoteck137/authlab/config"
	"github.com/nanoteck137/authlab/database"
	"github.com/nanoteck137/authlab/tools/utils"
	"golang.org/x/oauth2"
)

var (
	ErrAuthServiceSessionAlreadyExists = errors.New("AuthService: session already exists")
	ErrAuthServiceSessionNotFound      = errors.New("AuthService: session not found")
	ErrAuthServiceSessionExpired       = errors.New("AuthService: session is expired")
)

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
	Id     string
	Type   AuthSessionType
	Status AuthSessionStatus

	OAuth2Code string

	Expires time.Time
	Delete  time.Time
}

type AuthService struct {
	db        *database.Database
	jwtSecret string

	initialized bool

	Sessions map[string]*AuthSession

	provider     *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

func NewAuthService(db *database.Database, jwtSecret string) *AuthService {
	return &AuthService{
		db:          db,
		jwtSecret:   jwtSecret,
		initialized: false,
		Sessions:    make(map[string]*AuthSession),
	}
}

type SessionResult struct {
	SessionId string
	AuthUrl   string
	Expires   time.Time
}

func (a *AuthService) CreateNormalSession() (SessionResult, error) {
	// TODO(patrik): Add init check?

	id := utils.CreateId()

	t := time.Now()
	session := &AuthSession{
		Id:      id,
		Type:    AuthSessionTypeNormal,
		Status:  AuthSessionStatusPending,
		Expires: t.Add(5 * time.Minute),
		Delete:  t.Add(1 * time.Hour),
	}

	_, exists := a.Sessions[id]
	if exists {
		return SessionResult{}, ErrAuthServiceSessionAlreadyExists
	}

	a.Sessions[id] = session

	authUrl := a.oauth2Config.AuthCodeURL(session.Id)

	return SessionResult{
		SessionId: id,
		AuthUrl:   authUrl,
		Expires:   session.Expires,
	}, nil
}

func (a *AuthService) CompleteSession(sessionId, code string) error {
	session, exists := a.Sessions[sessionId]
	if !exists {
		return ErrAuthServiceSessionNotFound
	}

	if time.Now().After(session.Expires) {
		session.Status = AuthSessionStatusExpired
		return ErrAuthServiceSessionExpired
	}

	session.Status = AuthSessionStatusCompleted
	session.OAuth2Code = code

	return nil
}

func (a *AuthService) GetAuthCode(sessionId string) (*string, error) {
	session, exists := a.Sessions[sessionId]
	if !exists {
		return nil, ErrAuthServiceSessionNotFound
	}

	if session.Status != AuthSessionStatusCompleted {
		return nil, nil
	}

	return &session.OAuth2Code, nil
}

func (a *AuthService) GetUserFromCode(ctx context.Context, code string) (string, error) {
	oauth2Token, err := a.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return "", err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return "", errors.New("failed to login")
	}

	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", err
	}

	// {
	// 	var t map[string]any
	// 	err = idToken.Claims(&t)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	pretty.Println(t)
	// }

	// Extract user claims from OIDC token
	var oidcClaims struct {
		Email       string `json:"email"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Picture     string `json:"picture"`
		Sub         string `json:"sub"`
	}
	err = idToken.Claims(&oidcClaims)
	if err != nil {
		return "", err
	}

	pretty.Println(oidcClaims)

	getOrCreateUser := func() (string, error) {
		user, err := a.db.GetUserByEmail(ctx, oidcClaims.Email)
		if err != nil {
			if errors.Is(err, database.ErrItemNotFound) {
				user, err = a.db.CreateUser(ctx, database.CreateUserParams{
					Email:       oidcClaims.Email,
					DisplayName: oidcClaims.DisplayName,
					Role:        "user",
				})
				if err != nil {
					return "", err
				}

				return user.Id, nil
			} else {
				return "", err
			}
		}

		return user.Id, nil
	}

	identity, err := a.db.GetUserIdentity(ctx, "pocketid", oidcClaims.Sub)
	if err != nil {
		if errors.Is(err, database.ErrItemNotFound) {
			userId, err := getOrCreateUser()
			if err != nil {
				return "", err
			}

			err = a.db.CreateUserIdentity(ctx, database.CreateUserIdentityParams{
				Provider:   "pocketid",
				ProviderId: oidcClaims.Sub,
				UserId:     userId,
			})
			if err != nil {
				return "", err
			}

			return userId, nil
		} else {
			return "", err
		}
	}

	return identity.UserId, nil
}

func (a *AuthService) SignUserToken(userId string) (string, error) {
	user, err := a.db.GetUserById(context.Background(), userId)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.Id,
		"iat":    time.Now().Unix(),
		// "exp":    time.Now().Add(1000 * time.Second).Unix(),
	})

	tokenString, err := token.SignedString(([]byte)(a.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
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
		slog.Info("AuthService: running auth cleanup")
		a.RemoveUnusedEntries()
	}
}

func (a *AuthService) Init(ctx context.Context, config *config.Config) error {
	if a.initialized {
		return nil
	}

	var err error

	a.provider, err = oidc.NewProvider(ctx, config.OidcIssuerUrl)
	if err != nil {
		return fmt.Errorf("AuthService: failed to create OIDC provider: %w", err)
	}

	a.oauth2Config = &oauth2.Config{
		ClientID:     config.OidcClientId,
		ClientSecret: config.OidcClientSecret,
		RedirectURL:  config.OidcRedirectUrl,
		Endpoint:     a.provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	a.verifier = a.provider.Verifier(&oidc.Config{ClientID: config.OidcClientId})

	go a.RunRoutine()

	a.initialized = true

	return nil
}
