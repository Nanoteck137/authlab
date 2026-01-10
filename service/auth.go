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
	ErrAuthServiceRequestAlreadyExists = errors.New("AuthService: request already exists")
	ErrAuthServiceRequestNotFound      = errors.New("AuthService: request not found")
	ErrAuthServiceRequestExpired       = errors.New("AuthService: request is expired")
)

type AuthRequestType string

const (
	AuthRequestTypeNormal    AuthRequestType = "normal"
	AuthRequestTypeQuickCode AuthRequestType = "quick-code"
)

type AuthRequestStatus string

const (
	AuthRequestStatusPending   AuthRequestStatus = "pending"
	AuthRequestStatusCompleted AuthRequestStatus = "completed"
	AuthRequestStatusExpired   AuthRequestStatus = "expired"
	AuthRequestStatusFailed    AuthRequestStatus = "failed"
)

type AuthRequest struct {
	Id     string
	Type   AuthRequestType
	Status AuthRequestStatus

	OAuth2Code string

	Expires time.Time
	Delete  time.Time
}

type AuthService struct {
	db        *database.Database
	jwtSecret string

	initialized bool

	Requests map[string]*AuthRequest

	provider     *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

func NewAuthService(db *database.Database, jwtSecret string) *AuthService {
	return &AuthService{
		db:          db,
		jwtSecret:   jwtSecret,
		initialized: false,
		Requests:    make(map[string]*AuthRequest),
	}
}

type RequestResult struct {
	RequestId string
	AuthUrl   string
	Expires   time.Time
}

func (a *AuthService) CreateNormalRequest() (RequestResult, error) {
	// TODO(patrik): Add init check?

	id := utils.CreateId()

	t := time.Now()
	request := &AuthRequest{
		Id:      id,
		Type:    AuthRequestTypeNormal,
		Status:  AuthRequestStatusPending,
		Expires: t.Add(5 * time.Minute),
		Delete:  t.Add(1 * time.Hour),
	}

	_, exists := a.Requests[id]
	if exists {
		return RequestResult{}, ErrAuthServiceRequestAlreadyExists
	}

	a.Requests[id] = request

	authUrl := a.oauth2Config.AuthCodeURL(request.Id)

	return RequestResult{
		RequestId: id,
		AuthUrl:   authUrl,
		Expires:   request.Expires,
	}, nil
}

func (a *AuthService) CompleteRequest(requestId, code string) error {
	request, exists := a.Requests[requestId]
	if !exists {
		return ErrAuthServiceRequestNotFound
	}

	if time.Now().After(request.Expires) {
		request.Status = AuthRequestStatusExpired
		return ErrAuthServiceRequestExpired
	}

	request.Status = AuthRequestStatusCompleted
	request.OAuth2Code = code

	return nil
}

func (a *AuthService) GetAuthCode(requestId string) (*string, error) {
	request, exists := a.Requests[requestId]
	if !exists {
		return nil, ErrAuthServiceRequestNotFound
	}

	if request.Status != AuthRequestStatusCompleted {
		return nil, nil
	}

	return &request.OAuth2Code, nil
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
	for k, request := range a.Requests {
		if now.After(request.Delete) {
			delete(a.Requests, k)
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
