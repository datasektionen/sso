package oidcprovider

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	user "github.com/datasektionen/logout/services/user/export"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"golang.org/x/text/language"
)

// http://localhost:7000/op/authorize?client_id=bing&response_type=token&scope=openid&redirect_uri=http://localhost:8080/callback

type service struct {
	provider   *op.Provider
	user       user.Service
	db         *database.Queries
	dotabase   dotabase
	signingKey *ecdsa.PrivateKey
}

type dotabase struct {
	reqByID         map[uuid.UUID]authRequest
	reqIdByAuthCode map[string]uuid.UUID
	mu              sync.Mutex
}

var _ op.Storage = &service{}

func NewService(db *database.Queries) (*service, error) {
	var privateKey ecdsa.PrivateKey
	if err := json.Unmarshal([]byte(config.Config.OIDCProviderKey), &privateKey); err != nil {
		return nil, err
	}
	privateKey.Curve = elliptic.P256()
	s := &service{
		db: db,
		dotabase: dotabase{
			reqByID:         make(map[uuid.UUID]authRequest),
			reqIdByAuthCode: make(map[string]uuid.UUID),
		},
		signingKey: &privateKey,
	}
	var opts []op.Option
	if config.Config.Dev {
		opts = append(opts, op.WithAllowInsecure())
	}
	var key [32]byte

	if _, err := rand.Read(key[:]); err != nil {
		return nil, err
	}
	var err error
	s.provider, err = op.NewProvider(&op.Config{
		CryptoKey:          key,
		SupportedUILocales: []language.Tag{language.English},
		SupportedClaims: []string{
			"aud", "exp", "iat", "iss", "c_hash", "at_hash", "azp", // "scopes",
			"sub",
			"name", "family_name", "given_name",
			"email", "email_verified",
		},
	},
		s,
		op.StaticIssuer(config.Config.OIDCProviderIssuerURL.String()),
		opts...,
	)
	if err != nil {
		return nil, err
	}

	if config.Config.OIDCProviderIssuerURL.Path != "/op" {
		return nil, errors.New("The path of $OIDC_PROVIDER_ISSUER_URL must be `/`")
	}

	http.Handle("/op/", http.StripPrefix("/op", s.provider.Handler))
	http.Handle("/op-callback", httputil.Route(s.callback))

	return s, nil
}

func (s *service) Assign(user user.Service) {
	s.user = user
}

func (s *service) callback(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	authRequestID := r.FormValue("auth-request-id")
	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	s.dotabase.mu.Lock()
	defer s.dotabase.mu.Unlock()
	req, ok := s.dotabase.reqByID[id]
	if !ok {
		return httputil.BadRequest("No request with that id")
	}

	req.kthid, err = s.user.GetLoggedInKTHID(r)
	if err != nil {
		return err
	}
	if req.kthid == "" {
		return httputil.BadRequest("User did not seem to get logged in")
	}
	s.dotabase.reqByID[id] = req

	http.Redirect(w, r, "/op"+op.AuthCallbackURL(s.provider)(r.Context(), authRequestID), http.StatusSeeOther)
	return nil
}

// AuthRequestByCode implements op.Storage.
func (s *service) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	s.dotabase.mu.Lock()
	defer s.dotabase.mu.Unlock()
	id, ok := s.dotabase.reqIdByAuthCode[code]
	if !ok {
		return nil, httputil.BadRequest("Invalid code")
	}
	req, ok := s.dotabase.reqByID[id]
	if !ok {
		return nil, errors.New("Valid code but omg i lost the request")
	}
	return req, nil
}

// AuthRequestByID implements op.Storage.
func (s *service) AuthRequestByID(ctx context.Context, authRequestID string) (op.AuthRequest, error) {
	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return nil, httputil.BadRequest("Invalid uuid")
	}
	s.dotabase.mu.Lock()
	req, ok := s.dotabase.reqByID[id]
	s.dotabase.mu.Unlock()
	if !ok {
		return nil, httputil.BadRequest("No request with that id")
	}
	return req, nil
}

// AuthorizeClientIDSecret implements op.Storage.
func (s *service) AuthorizeClientIDSecret(ctx context.Context, clientID string, clientSecret string) error {
	slog.Warn("oidcprovider.*service.AuthorizeClientIDSecret", "clientID", clientID, "clientSecret", clientSecret)
	if clientID != "todo" || clientSecret != "secrettodo" {
		return httputil.BadRequest("Invalid client id or secret")
	}
	return nil
}

// CreateAccessAndRefreshTokens implements op.Storage.
func (s *service) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (accessTokenID string, newRefreshTokenID string, expiration time.Time, err error) {
	slog.Warn("oidcprovider.*service.CreateAccessAndRefreshTokens", "request", request, "currentRefreshToken", currentRefreshToken)
	panic("unimplemented")
}

// CreateAccessToken implements op.Storage.
func (s *service) CreateAccessToken(ctx context.Context, request op.TokenRequest) (accessTokenID string, expiration time.Time, err error) {
	slog.Warn("oidcprovider.*service.CreateAccessToken", "request", request)
	return "", time.Now().Add(time.Minute), nil
}

// CreateAuthRequest implements op.Storage.
func (s *service) CreateAuthRequest(ctx context.Context, r *oidc.AuthRequest, userID string) (op.AuthRequest, error) {
	if userID != "" {
		slog.Info("oidcprovider.*service.CreateAuthRequest: we got a userID!!!", "userID", userID)
	}

	id := uuid.New()
	req := authRequest{id: id, authCode: "", inner: r}
	s.dotabase.mu.Lock()
	s.dotabase.reqByID[id] = req
	s.dotabase.mu.Unlock()
	return req, nil
}

// DeleteAuthRequest implements op.Storage.
func (s *service) DeleteAuthRequest(ctx context.Context, authRequestID string) error {
	slog.Warn("oidcprovider.*service.DeleteAuthRequest", "authRequestID", authRequestID)

	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	s.dotabase.mu.Lock()
	defer s.dotabase.mu.Unlock()
	req, ok := s.dotabase.reqByID[id]
	if !ok {
		return httputil.BadRequest("No request with that id")
	}
	delete(s.dotabase.reqByID, id)
	delete(s.dotabase.reqIdByAuthCode, req.authCode)
	return nil
}

// GetClientByClientID implements op.Storage.
func (s *service) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
	slog.Warn("oidcprovider.*service.GetClientByClientID", "clientID", clientID)
	return client{id: "todo"}, nil
}

// GetKeyByIDAndClientID implements op.Storage.
func (s *service) GetKeyByIDAndClientID(ctx context.Context, keyID string, clientID string) (*jose.JSONWebKey, error) {
	slog.Warn("oidcprovider.*service.GetKeyByIDAndClientID", "keyID", keyID, "clientID", clientID)
	panic("unimplemented")
}

// GetPrivateClaimsFromScopes implements op.Storage.
func (s *service) GetPrivateClaimsFromScopes(ctx context.Context, userID string, clientID string, scopes []string) (map[string]any, error) {
	slog.Warn("oidcprovider.*service.GetPrivateClaimsFromScopes", "userID", userID, "clientID", clientID, "scopes", scopes)
	panic("unimplemented")
}

// GetRefreshTokenInfo implements op.Storage.
func (s *service) GetRefreshTokenInfo(ctx context.Context, clientID string, token string) (userID string, tokenID string, err error) {
	slog.Warn("oidcprovider.*service.GetRefreshTokenInfo", "clientID", clientID, "token", token)
	panic("unimplemented")
}

// Health implements op.Storage.
func (s *service) Health(ctx context.Context) error {
	slog.Warn("oidcprovider.*service.Health")
	panic("unimplemented")
}

// RevokeToken implements op.Storage.
func (s *service) RevokeToken(ctx context.Context, tokenOrTokenID string, userID string, clientID string) *oidc.Error {
	slog.Warn("oidcprovider.*service.RevokeToken", "tokenOrTokenID", tokenOrTokenID, "userID", userID, "clientID", clientID)
	panic("unimplemented")
}

// SaveAuthCode implements op.Storage.
func (s *service) SaveAuthCode(ctx context.Context, authRequestID string, code string) error {
	slog.Warn("oidcprovider.*service.SaveAuthCode", "authRequestID", authRequestID, "code", code)
	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	s.dotabase.mu.Lock()
	defer s.dotabase.mu.Unlock()
	req, ok := s.dotabase.reqByID[id]
	if !ok {
		return httputil.BadRequest("No request with that id")
	}
	s.dotabase.reqIdByAuthCode[code] = id
	req.authCode = code
	s.dotabase.reqByID[id] = req
	return nil
}

// SetIntrospectionFromToken implements op.Storage.
func (s *service) SetIntrospectionFromToken(ctx context.Context, userinfo *oidc.IntrospectionResponse, tokenID string, subject string, clientID string) error {
	slog.Warn("oidcprovider.*service.SetIntrospectionFromToken", "userinfo", userinfo, "tokenID", tokenID, "subject", subject, "clientID", clientID)
	panic("unimplemented")
}

// SetUserinfoFromScopes implements op.Storage.
func (s *service) SetUserinfoFromScopes(ctx context.Context, userinfo *oidc.UserInfo, kthid string, clientID string, scopes []string) error {
	slog.Error("oidcprovider.*service.SetUserinfoFromScopes", "kthid", kthid, "clientID", clientID, "scopes", scopes)
	user, err := s.user.GetUser(ctx, kthid)
	if err != nil {
		return err
	}
	if user == nil {
		slog.Error("SetUserinfoFromScopes: user not found", "kthid", kthid, "clientID", clientID, "scopes", scopes)
		panic("SetUserinfoFromScopes, no user but pretty sure that should have been handled in this request???")
	}
	for _, scope := range scopes {
		switch scope {
		case oidc.ScopeOpenID:
			userinfo.Subject = kthid
		case oidc.ScopeProfile:
			userinfo.GivenName = user.FirstName
			userinfo.FamilyName = user.Surname
		case oidc.ScopeEmail:
			userinfo.Email = user.Email
			userinfo.EmailVerified = true
		}
	}
	return nil
}

// SetUserinfoFromToken implements op.Storage.
func (s *service) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID string, kthid string, origin string) error {
	slog.Warn("oidcprovider.*service.SetUserinfoFromToken", "userinfo", userinfo, "tokenID", tokenID, "subject", kthid, "origin", origin)
	userinfo.Subject = kthid
	return nil
}

// SignatureAlgorithms implements op.Storage.
func (s *service) SignatureAlgorithms(ctx context.Context) ([]jose.SignatureAlgorithm, error) {
	return []jose.SignatureAlgorithm{jose.ES256}, nil
}

type publicKey struct{ *ecdsa.PublicKey }

func (s publicKey) ID() string                         { return "the-one-and-only" }
func (s publicKey) Key() any                           { return s.PublicKey }
func (s publicKey) Algorithm() jose.SignatureAlgorithm { return jose.ES256 }
func (s publicKey) Use() string                        { return "sig" }

var _ op.Key = publicKey{}

// KeySet implements op.Storage.
func (s *service) KeySet(ctx context.Context) ([]op.Key, error) {
	return []op.Key{publicKey{&s.signingKey.PublicKey}}, nil
}

type privateKey struct{ *ecdsa.PrivateKey }

func (s privateKey) ID() string                                  { return "the-one-and-only" }
func (s privateKey) Key() any                                    { return s.PrivateKey }
func (s privateKey) SignatureAlgorithm() jose.SignatureAlgorithm { return jose.ES256 }

var _ op.SigningKey = privateKey{}

// SigningKey implements op.Storage.
func (s *service) SigningKey(ctx context.Context) (op.SigningKey, error) {
	return privateKey{s.signingKey}, nil
}

// TerminateSession implements op.Storage.
func (s *service) TerminateSession(ctx context.Context, userID string, clientID string) error {
	slog.Warn("oidcprovider.*service.TerminateSession", "userID", userID, "clientID", clientID)
	panic("unimplemented")
}

// TokenRequestByRefreshToken implements op.Storage.
func (s *service) TokenRequestByRefreshToken(ctx context.Context, refreshTokenID string) (op.RefreshTokenRequest, error) {
	slog.Warn("oidcprovider.*service.TokenRequestByRefreshToken", "refreshTokenID", refreshTokenID)
	panic("unimplemented")
}

// ValidateJWTProfileScopes implements op.Storage.
func (s *service) ValidateJWTProfileScopes(ctx context.Context, userID string, scopes []string) ([]string, error) {
	slog.Warn("oidcprovider.*service.ValidateJWTProfileScopes", "userID", userID, "scopes", scopes)
	panic("unimplemented")
}
