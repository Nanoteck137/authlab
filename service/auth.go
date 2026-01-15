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

const (
	AuthRequestExpireDuration   = 5 * time.Minute
	AuthRequestDeletionDuration = 1 * time.Hour
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
	Id         string
	ProviderId string

	Type   AuthRequestType
	Status AuthRequestStatus

	OAuth2Code string

	Expires time.Time
	Delete  time.Time
}

type AuthProvider struct {
	Id          string
	DisplayName string

	provider     *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

func (p *AuthProvider) init(ctx context.Context, config config.ConfigOidcProvider) error {
	var err error

	p.provider, err = oidc.NewProvider(ctx, config.IssuerUrl)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider (%s): %w", err, p.Id)
	}

	p.oauth2Config = &oauth2.Config{
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectUrl,
		Endpoint:     p.provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	p.verifier = p.provider.Verifier(&oidc.Config{ClientID: config.ClientId})

	return nil
}

type providerClaim struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Picture     string `json:"picture"`
	Sub         string `json:"sub"`
}

func (p *AuthProvider) claim(ctx context.Context, code string) (providerClaim, error) {
	oauth2Token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return providerClaim{}, err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return providerClaim{}, errors.New("failed to login")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return providerClaim{}, err
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

	var claims providerClaim
	err = idToken.Claims(&claims)
	if err != nil {
		return providerClaim{}, err
	}

	return claims, nil
}

type AuthService struct {
	db        *database.Database
	jwtSecret string

	initialized bool

	providers map[string]*AuthProvider

	Requests map[string]*AuthRequest
}

func NewAuthService(db *database.Database, jwtSecret string) *AuthService {
	return &AuthService{
		db:          db,
		jwtSecret:   jwtSecret,
		initialized: false,
		providers:   make(map[string]*AuthProvider),
		Requests:    make(map[string]*AuthRequest),
	}
}

type RequestResult struct {
	RequestId string
	AuthUrl   string
	Expires   time.Time
}

func (a *AuthService) CreateNormalRequest(providerId string) (RequestResult, error) {
	// TODO(patrik): Add init check?

	provider, exists := a.providers[providerId]
	if !exists {
		// TODO(patrik): Better error
		return RequestResult{}, errors.New("provider not found")
	}

	id := utils.CreateId()

	t := time.Now()
	request := &AuthRequest{
		Id:         id,
		ProviderId: provider.Id,
		Type:       AuthRequestTypeNormal,
		Status:     AuthRequestStatusPending,
		Expires:    t.Add(AuthRequestExpireDuration),
		Delete:     t.Add(AuthRequestDeletionDuration),
	}

	_, exists = a.Requests[id]
	if exists {
		return RequestResult{}, ErrAuthServiceRequestAlreadyExists
	}

	a.Requests[id] = request

	authUrl := provider.oauth2Config.AuthCodeURL(request.Id)

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

	if request.Status == AuthRequestStatusPending {
		request.Status = AuthRequestStatusCompleted
		request.OAuth2Code = code
	}

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

func (a *AuthService) InvalidateRequest(requestId string) error {
	request, exists := a.Requests[requestId]
	if !exists {
		return ErrAuthServiceRequestNotFound
	}

	request.Status = AuthRequestStatusExpired

	return nil
}

func (a *AuthService) GetUserFromCode(ctx context.Context, providerId, code string) (string, error) {
	provider, exists := a.providers[providerId]
	if !exists {
		// TODO(patrik): Fix error
		return "", errors.New("provider not found")
	}

	oidcClaims, err := provider.claim(ctx, code)
	if err != nil {
		return "", err
	}

	pretty.Println(oidcClaims)

	getOrCreateUser := func() (string, error) {
		user, err := a.db.GetUserByEmail(ctx, oidcClaims.Email)
		if err != nil {
			if errors.Is(err, database.ErrItemNotFound) {
				displayName := oidcClaims.DisplayName
				if displayName == "" {
					displayName = oidcClaims.Name
				}

				user, err = a.db.CreateUser(ctx, database.CreateUserParams{
					Email:       oidcClaims.Email,
					DisplayName: displayName,
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

	identity, err := a.db.GetUserIdentity(ctx, providerId, oidcClaims.Sub)
	if err != nil {
		if errors.Is(err, database.ErrItemNotFound) {
			userId, err := getOrCreateUser()
			if err != nil {
				return "", err
			}

			err = a.db.CreateUserIdentity(ctx, database.CreateUserIdentityParams{
				Provider:   providerId,
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

	for id, provider := range config.OidcProviders {
		res := &AuthProvider{
			Id:          id,
			DisplayName: provider.Name,
		}

		// TODO(patrik): Store the error on the provider and show
		// that in the frontend
		err := res.init(ctx, provider)
		if err != nil {
			return fmt.Errorf("AuthService: failed to initialize AuthProvider: %w", err)
		}

		a.providers[id] = res
	}

	go a.RunRoutine()

	a.initialized = true

	return nil
}
