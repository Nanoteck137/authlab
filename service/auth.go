package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nanoteck137/authlab/config"
	"github.com/nanoteck137/authlab/database"
	"github.com/nanoteck137/authlab/tools/utils"
	"golang.org/x/oauth2"
)

type ServiceError struct {
	Service string
	Err     error
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Service, e.Err)
}

type ServiceErrCreator struct {
	Service string
}

func NewServiceErrCreator(service string) ServiceErrCreator {
	return ServiceErrCreator{
		Service: service,
	}
}

func (s *ServiceErrCreator) Error(text string) error {
	return &ServiceError{
		Service: s.Service,
		Err:     errors.New(text),
	}
}

func (s *ServiceErrCreator) Errorf(format string, a ...any) error {
	return &ServiceError{
		Service: s.Service,
		Err:     fmt.Errorf(format, a...),
	}
}

var authErr = NewServiceErrCreator("auth-service")

var (
	ErrAuthServiceProviderNotFound = authErr.Error("provider not found")

	ErrAuthServiceRequestAlreadyExists = authErr.Error("request already exists")
	ErrAuthServiceRequestNotFound      = authErr.Error("request not found")
	ErrAuthServiceRequestExpired       = authErr.Error("request is expired")
	ErrAuthServiceRequestNotReady      = authErr.Error("request is not ready")
	ErrAuthServiceRequestInvalid       = authErr.Error("request is invalid")
)

const (
	authProviderRequestExpireDuration   = 5 * time.Minute
	authProviderRequestDeletionDuration = authProviderRequestExpireDuration + 10*time.Minute

	authQuickRequestExpireDuration   = 5 * time.Minute
	authQuickRequestDeletionDuration = authQuickRequestExpireDuration + 10*time.Minute
)

type AuthProviderRequestStatus string

const (
	AuthProviderRequestStatusPending   AuthProviderRequestStatus = "pending"
	AuthProviderRequestStatusCompleted AuthProviderRequestStatus = "completed"
	AuthProviderRequestStatusExpired   AuthProviderRequestStatus = "expired"
	AuthProviderRequestStatusFailed    AuthProviderRequestStatus = "failed"
)

type AuthQuickRequestStatus string

const (
	AuthQuickRequestStatusPending   AuthQuickRequestStatus = "pending"
	AuthQuickRequestStatusCompleted AuthQuickRequestStatus = "completed"
	AuthQuickRequestStatusExpired   AuthQuickRequestStatus = "expired"
	AuthQuickRequestStatusFailed    AuthQuickRequestStatus = "failed"
)

// authProviderRequest holds the infomation for a provider auth request
type authProviderRequest struct {
	// Unique id of this request
	id         string

	// Id of the provider this request belongs to
	providerId string

	// Status of the request
	status AuthProviderRequestStatus

	// Challenge code used to check when trying to do something with 
	// the request
	challenge string

	// The generated provider url saved for later use
	oauth2Url  string

	// The code the provider sends back inside the callback, this should be 
	// set when status is completed and then we can generate the user 
	// token based on the code
	oauth2Code string

	// The timestamp for when this request is invalid
	expires time.Time

	// The timestamp for when this request can be deleted
	delete time.Time
}

// authProvider hold infomation about the OAuth2/OIDC provider
type authProvider struct {
	// If the provider has been initialized
	initialized bool

	// The id of the provider
	id          string

	// A display name for showing in the frontend
	displayName string

	// The config object that holds infomation used by the provider
	config config.ConfigOidcProvider

	// The OIDC provider object
	provider     *oidc.Provider

	// The OAuth2 config object
	oauth2Config *oauth2.Config

	// The OIDC token verifier
	verifier     *oidc.IDTokenVerifier
}

func (p *authProvider) init(ctx context.Context) error {
	if p.initialized {
		return nil
	}

	var err error

	p.provider, err = oidc.NewProvider(ctx, p.config.IssuerUrl)
	if err != nil {
		return err
	}

	p.oauth2Config = &oauth2.Config{
		ClientID:     p.config.ClientId,
		ClientSecret: p.config.ClientSecret,
		RedirectURL:  p.config.RedirectUrl,
		Endpoint:     p.provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	p.verifier = p.provider.Verifier(&oidc.Config{ClientID: p.config.ClientId})

	p.initialized = true

	return nil
}

type providerClaim struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Picture     string `json:"picture"`
	Sub         string `json:"sub"`
}

func (p *authProvider) claim(ctx context.Context, code string) (providerClaim, error) {
	oauth2Token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return providerClaim{}, err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return providerClaim{}, errors.New("oauth2 token is missing id_token")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return providerClaim{}, err
	}

	var claims providerClaim
	err = idToken.Claims(&claims)
	if err != nil {
		return providerClaim{}, err
	}

	return claims, nil
}

type authQuickConnectRequest struct {
	// the status of the request
	status AuthQuickRequestStatus

	// the code of this request, generated by GenerateCode
	// example "ABCD-1234-EFGH"
	code string

	// small token generated by GenerateAuthChallenge and is returned
	// to the client when initiating, then is checked when trying to
	// create the JWT token
	challenge string

	// This is set when status is completed, and is the userId of
	// the user that authorized the request
	userId string

	// the expiry date of this request
	expires time.Time

	// the date of then we can delete this request
	delete time.Time
}

type AuthService struct {
	// The lock for the service, needs to be claimed when modifing request data
	mu sync.Mutex

	// Reference to the database
	db *database.Database

	// the jwt secret used to sign user tokens
	jwtSecret string

	// The available providers
	providers map[string]*authProvider

	// All the provider based requests
	ProviderRequests map[string]*authProviderRequest

	// All the quick connect based requests
	QuickConnectRequests map[string]*authQuickConnectRequest
}

func NewAuthService(db *database.Database, config *config.Config) *AuthService {
	providers := make(map[string]*authProvider, len(config.OidcProviders))

	for id, providerConfig := range config.OidcProviders {
		res := &authProvider{
			id:          id,
			displayName: providerConfig.Name,
			config:      providerConfig,
		}

		providers[id] = res
	}

	return &AuthService{
		db:                   db,
		jwtSecret:            config.JwtSecret,
		providers:            providers,
		ProviderRequests:     make(map[string]*authProviderRequest),
		QuickConnectRequests: make(map[string]*authQuickConnectRequest),
	}
}

// ProviderRequestResult is the structure returned by CreateProviderRequest
// and contains some data about the newly created request
type ProviderRequestResult struct {
	// The Request Id
	RequestId string

	// The Auth URL used to redirect/open window so that the user
	// can authenticate with the provider
	AuthUrl string

	// The challenge code used to varify request calls
	Challenge string

	// The timestamp when this request expires
	Expires time.Time
}

// CreateProviderRequest creates a provider request and returns some
// data about the request so that the user can complete the request
func (a *AuthService) CreateProviderRequest(providerId string) (ProviderRequestResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check for the provider
	provider, exists := a.providers[providerId]
	if !exists {
		return ProviderRequestResult{}, ErrAuthServiceProviderNotFound
	}

	// Initialize the provider, this checks for if it's already initialized
	err := provider.init(context.TODO())
	if err != nil {
		return ProviderRequestResult{}, authErr.Errorf("initialize AuthProvider(%s): %w", provider.id, err)
	}

	// Generate the unique challenge used to check calls to requests
	challenge, err := utils.GenerateAuthChallenge()
	if err != nil {
		return ProviderRequestResult{}, authErr.Errorf("generate auth challenge: %w", err)
	}

	// Create a unique id for the request
	id := utils.CreateId()

	// Create the request
	t := time.Now()
	request := &authProviderRequest{
		id:         id,
		providerId: provider.id,
		status:     AuthProviderRequestStatusPending,
		challenge:  challenge,
		expires:    t.Add(authProviderRequestExpireDuration),
		delete:     t.Add(authProviderRequestDeletionDuration),
	}

	// Generate the OAuth2 URL so that the frontend can redirect/open window
	// with this url
	request.oauth2Url = provider.oauth2Config.AuthCodeURL(request.id)

	// Check if the request id is already used
	_, exists = a.ProviderRequests[id]
	if exists {
		return ProviderRequestResult{}, ErrAuthServiceRequestAlreadyExists
	}

	// Save the request
	a.ProviderRequests[id] = request

	return ProviderRequestResult{
		RequestId: request.id,
		AuthUrl:   request.oauth2Url,
		Challenge: request.challenge,
		Expires:   request.expires,
	}, nil
}

// QuickConnectRequestResult is the structure returned by
// CreateQuickConnectRequest and contains some data about the newly
// created request
type QuickConnectRequestResult struct {
	// The Quick Connect code the user need to use for claiming
	Code string

	// The challenge code used to varify request calls
	Challenge string

	// The timestamp when this request expires
	Expires time.Time
}

// CreateQuickConnectRequest creates a quick connect request and returns some
// data about the request so that the user can complete the request
func (a *AuthService) CreateQuickConnectRequest() (QuickConnectRequestResult, error) {
	// Generate the unique code for this quick connect request
	code, err := utils.GenerateCode()
	if err != nil {
		return QuickConnectRequestResult{}, fmt.Errorf("failed to generate code: %w", err)
	}

	// Generate the unique challenge used to check calls to requests
	challenge, err := utils.GenerateAuthChallenge()
	if err != nil {
		return QuickConnectRequestResult{}, fmt.Errorf("failed to generate challenge: %w", err)
	}

	// Create the request
	t := time.Now()
	request := &authQuickConnectRequest{
		status:    AuthQuickRequestStatusPending,
		code:      code,
		challenge: challenge,
		expires:   t.Add(authQuickRequestExpireDuration),
		delete:    t.Add(authQuickRequestDeletionDuration),
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if the code is already used
	_, exists := a.QuickConnectRequests[code]
	if exists {
		return QuickConnectRequestResult{}, err
	}

	// Save the request
	a.QuickConnectRequests[code] = request

	return QuickConnectRequestResult{
		Code:      request.code,
		Challenge: request.challenge,
		Expires:   request.expires,
	}, nil
}

// CompleteProviderRequest this is called after a user claims this
// quick connect request after this we then can create the user token with
// a call to CreateAuthTokenForQuickConnect
func (a *AuthService) CompleteQuickConnectRequest(requestCode, userId string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get the request
	request, exists := a.QuickConnectRequests[requestCode]
	if !exists {
		return ErrAuthServiceRequestNotFound
	}

	// Check if the request is expired and update the status
	if time.Now().After(request.expires) {
		request.status = AuthQuickRequestStatusExpired
		return ErrAuthServiceRequestExpired
	}

	// Check if the request is pending and update the status +
	// save the userId
	if request.status == AuthQuickRequestStatusPending {
		request.status = AuthQuickRequestStatusCompleted
		request.userId = userId
	}

	return nil
}

// CompleteProviderRequest this is called after we get the code from
// the OAuth2 provider and with the code we now can set the request status
// to completed and later call CreateAuthTokenForProvider to
// generate the user token
func (a *AuthService) CompleteProviderRequest(requestId, code string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get the request
	request, exists := a.ProviderRequests[requestId]
	if !exists {
		return ErrAuthServiceRequestNotFound
	}

	// Check if the request is expired and update the request status
	// if it is expired
	if time.Now().After(request.expires) {
		request.status = AuthProviderRequestStatusExpired
		return ErrAuthServiceRequestExpired
	}

	// Check for if the request status is pending and then set the status to
	// completed and saves the code for later use
	if request.status == AuthProviderRequestStatusPending {
		request.status = AuthProviderRequestStatusCompleted
		request.oauth2Code = code
	}

	return nil
}

// CheckProviderRequestStatus checks the request for if it's expired
// and then returns the current status of the request
func (a *AuthService) CheckProviderRequestStatus(requestId, challenge string) (AuthProviderRequestStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get the request
	request, exists := a.ProviderRequests[requestId]
	if !exists {
		return AuthProviderRequestStatusFailed, ErrAuthServiceRequestNotFound
	}

	// Test the challenge
	if request.challenge != challenge {
		// TODO(patrik): Give this it's own error
		return AuthProviderRequestStatusFailed, ErrAuthServiceRequestNotFound
	}

	// Check if the request is expired, and if it is set the request
	// status to expired
	now := time.Now()
	if now.After(request.expires) {
		request.status = AuthProviderRequestStatusExpired
	}

	return request.status, nil
}

// CheckQuickConnectRequestStatus checks the request for if it's expired
// and then returns the current status of the request
func (a *AuthService) CheckQuickConnectRequestStatus(requestCode, challenge string) (AuthQuickRequestStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get the request
	request, exists := a.QuickConnectRequests[requestCode]
	if !exists {
		return AuthQuickRequestStatusFailed, ErrAuthServiceRequestNotFound
	}

	// Test the challenge
	if request.challenge != challenge {
		return AuthQuickRequestStatusFailed, ErrAuthServiceRequestNotFound
	}

	// Check if the request is expired, and if it is set the request
	// status to expired
	now := time.Now()
	if now.After(request.expires) {
		request.status = AuthQuickRequestStatusExpired
	}

	return request.status, nil
}

// CreateAuthTokenForProvider create a user JWT token if the quick connect
// request is complete, otherwise return error
//
// Thread-safe: locks the service
func (a *AuthService) CreateAuthTokenForProvider(requestId, challenge string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get the request
	request, exists := a.ProviderRequests[requestId]
	if !exists {
		return "", ErrAuthServiceRequestNotFound
	}

	// Test the challenge
	if request.challenge != challenge {
		return "", ErrAuthServiceRequestNotFound
	}

	// Check the request status for completed
	if request.status != AuthProviderRequestStatusCompleted {
		return "", ErrAuthServiceRequestNotReady
	}

	// TODO(patrik): Check provider?
	// Get the provider from the request
	provider := a.providers[request.providerId]

	// Check if the OAuth2Code is set, this should be set after the
	// OAuth2 callback
	if request.oauth2Code == "" {
		// Set the request status to failed, because we have
		// encountered an error with the OAuth2 code
		request.status = AuthProviderRequestStatusFailed
		return "", ErrAuthServiceRequestInvalid
	}

	// Get the user id from the OAuth2 Code that is stored in the request
	// after the provider completes the OAuth2 request
	userId, err := a.getUserFromCode(context.TODO(), provider, request.oauth2Code)
	if err != nil {
		// Set the request status to failed, because we have
		// encountered an error with getting the user from the provider
		request.status = AuthProviderRequestStatusFailed
		return "", err
	}

	// Check the user id just to be sure, this is not 100% necessary
	// because the SignUserToken should check if the user is already
	// in the database
	if userId == "" {
		// Set the request status to failed, because we have
		// encountered an error with getting the user from the provider
		request.status = AuthProviderRequestStatusFailed
		return "", ErrAuthServiceRequestInvalid
	}

	// Set the request status to be expired so that we can't generate
	// the token after this
	request.status = AuthProviderRequestStatusExpired

	// Create the JWT token for the user
	token, err := a.SignUserToken(userId)
	if err != nil {
		request.status = AuthProviderRequestStatusFailed
		return "", err
	}

	return token, nil
}

// CreateAuthTokenForQuickConnect create a user JWT token if the quick connect
// request is complete, otherwise return error
//
// Thread-safe: locks the service
func (a *AuthService) CreateAuthTokenForQuickConnect(requestCode, challenge string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get the request
	request, exists := a.QuickConnectRequests[requestCode]
	if !exists {
		return "", ErrAuthServiceRequestNotFound
	}

	// Test the challenge
	if request.challenge != challenge {
		return "", ErrAuthServiceRequestNotFound
	}

	// Check the request status for completed
	if request.status != AuthQuickRequestStatusCompleted {
		return "", ErrAuthServiceRequestInvalid
	}

	// Check if the userId is set because if the request is completed then
	// this should be set after a user claims this request
	if request.userId == "" {
		return "", ErrAuthServiceRequestInvalid
	}

	// Set the request status to be expired so that we can't generate
	// the token after this
	request.status = AuthQuickRequestStatusExpired

	// Create the JWT token for the user
	token, err := a.SignUserToken(request.userId)
	if err != nil {
		return "", err
	}

	return token, nil
}

// getUserFromCode tries to returns the user id after claiming the OAuth2 code
func (a *AuthService) getUserFromCode(ctx context.Context, provider *authProvider, code string) (string, error) {
	oidcClaims, err := provider.claim(ctx, code)
	if err != nil {
		return "", authErr.Errorf("provider claim: %w", err)
	}

	// Helper function to get/create the user
	getOrCreateUser := func() (string, error) {
		// Check if the user with the email already exists
		user, err := a.db.GetUserByEmail(ctx, oidcClaims.Email)
		// If the user exists, just return that user id
		if err == nil {
			return user.Id, nil
		}

		// If the user doesn't exist we need to create a new user using
		// the oidcClaims
		if errors.Is(err, database.ErrItemNotFound) {
			// Try getting a display name for the user
			displayName := oidcClaims.DisplayName
			// If the provider doesn't provide the display name then just
			// use the name field
			if displayName == "" {
				displayName = oidcClaims.Name
			}

			// Create the database entry for the user
			user, err = a.db.CreateUser(ctx, database.CreateUserParams{
				Email:       oidcClaims.Email,
				DisplayName: displayName,
				Role:        "user",
			})
			if err != nil {
				return "", authErr.Errorf("create user: %w", err)
			}

			return user.Id, nil
		} else {
			return "", authErr.Errorf("get user by email: %w", err)
		}
	}

	identity, err := a.db.GetUserIdentity(ctx, provider.id, oidcClaims.Sub)
	// If no error, just return the user id
	if err == nil {
		return identity.UserId, nil
	}

	// Check for if the error is ErrItemNotFound, if it
	// is then create the identity entry
	if errors.Is(err, database.ErrItemNotFound) {
		// Try to get/create the user
		userId, err := getOrCreateUser()
		if err != nil {
			return "", err
		}

		// Then create the user identity entry
		err = a.db.CreateUserIdentity(ctx, database.CreateUserIdentityParams{
			Provider:   provider.id,
			ProviderId: oidcClaims.Sub,
			UserId:     userId,
		})
		if err != nil {
			return "", authErr.Errorf("create user identity: %w", err)
		}

		return userId, nil
	} else {
		return "", authErr.Errorf("get user identity: %w", err)
	}
}

// SignUserToken generates a JWT token for the giving user id.
// Returns the JWT token or error if the user doesn't exist or signing fails.
func (a *AuthService) SignUserToken(userId string) (string, error) {
	// Check if the user with the id exists in the database
	user, err := a.db.GetUserById(context.Background(), userId)
	if err != nil {
		return "", authErr.Errorf("signing token: get user by id: %w", err)
	}

	// Create jwt token with the for the user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.Id,
		"iat":    time.Now().Unix(),
		// "exp":    time.Now().Add(1000 * time.Second).Unix(),
	})

	// Sign the newly created token
	tokenString, err := token.SignedString(([]byte)(a.jwtSecret))
	if err != nil {
		return "", authErr.Errorf("signing token: jwt sign: %w", err)
	}

	return tokenString, nil
}

// RemoveUnusedEntries performs cleanup on expired auth requests
func (a *AuthService) RemoveUnusedEntries() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()

	// Remove expired provider OAuth2 requests
	for k, request := range a.ProviderRequests {
		if now.After(request.delete) {
			delete(a.ProviderRequests, k)
		}
	}

	// Remove expired quick connect requests
	for k, request := range a.QuickConnectRequests {
		if now.After(request.delete) {
			delete(a.QuickConnectRequests, k)
		}
	}
}

// TODO(patrik): This should be a worker that the app creates when initializing
func (a *AuthService) CleanRoutine() {
	ticker := time.NewTicker(30 * time.Minute)
	for range ticker.C {
		slog.Info("auth-service: running cleanup")
		a.RemoveUnusedEntries()
	}
}
