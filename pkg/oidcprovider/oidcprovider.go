package oidcprovider

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/service"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"golang.org/x/text/language"
)

// http://localhost:7000/op/authorize?client_id=bing&response_type=token&scope=openid&redirect_uri=http://localhost:8080/callback

type provider struct {
	provider   *op.Provider
	dotabase   dotabase
	signingKey *ecdsa.PrivateKey
	s          *service.Service
}

type dotabase struct {
	reqByID         map[uuid.UUID]authRequest
	reqIdByAuthCode map[string]uuid.UUID
	mu              sync.Mutex
}

var _ op.Storage = &provider{}

func MountRoutes(s *service.Service) error {
	privateKey := ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int),
			Y:     new(big.Int),
		},
		D: new(big.Int),
	}
	parts := strings.SplitN(config.Config.OIDCProviderKey, ",", 3)
	if _, ok := privateKey.X.SetString(parts[0], 62); !ok {
		return errors.New("Invalid x in $OIDC_PROVIDER_KEY")
	}
	if _, ok := privateKey.Y.SetString(parts[1], 62); !ok {
		return errors.New("Invalid y in $OIDC_PROVIDER_KEY")
	}
	if _, ok := privateKey.D.SetString(parts[2], 62); !ok {
		return errors.New("Invalid d in $OIDC_PROVIDER_KEY")
	}

	p := &provider{
		dotabase: dotabase{
			reqByID:         make(map[uuid.UUID]authRequest),
			reqIdByAuthCode: make(map[string]uuid.UUID),
		},
		signingKey: &privateKey,
		s:          s,
	}
	var opts []op.Option
	if config.Config.Dev {
		opts = append(opts, op.WithAllowInsecure())
	}
	var key [32]byte

	if _, err := rand.Read(key[:]); err != nil {
		return err
	}
	var err error
	p.provider, err = op.NewProvider(&op.Config{
		CryptoKey:          key,
		SupportedUILocales: []language.Tag{language.English},
		SupportedClaims: []string{
			"aud", "exp", "iat", "iss", "c_hash", "at_hash", "azp", // "scopes",
			"sub",
			"name", "family_name", "given_name",
			"email", "email_verified",
		},
	},
		p,
		op.StaticIssuer(config.Config.OIDCProviderIssuerURL.String()),
		opts...,
	)
	if err != nil {
		return err
	}

	if config.Config.OIDCProviderIssuerURL.Path != "/op" {
		return errors.New("The path of $OIDC_PROVIDER_ISSUER_URL must be `/`")
	}

	http.Handle("/op/", http.StripPrefix("/op", p.provider.Handler))
	http.Handle("/op-callback", httputil.Route(nil, p.callback))

	return nil
}

func (p *provider) callback(_ any, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	authRequestID := r.FormValue("auth-request-id")
	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	p.dotabase.mu.Lock()
	defer p.dotabase.mu.Unlock()
	req, ok := p.dotabase.reqByID[id]
	if !ok {
		return httputil.BadRequest("No request with that id")
	}

	req.kthid, err = p.s.GetLoggedInKTHID(r)
	if err != nil {
		return err
	}
	if req.kthid == "" {
		return httputil.BadRequest("User did not seem to get logged in")
	}
	p.dotabase.reqByID[id] = req

	http.Redirect(w, r, "/op"+op.AuthCallbackURL(p.provider)(r.Context(), authRequestID), http.StatusSeeOther)
	return nil
}

// AuthRequestByCode implements op.Storage.
func (p *provider) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	p.dotabase.mu.Lock()
	defer p.dotabase.mu.Unlock()
	id, ok := p.dotabase.reqIdByAuthCode[code]
	if !ok {
		return nil, httputil.BadRequest("Invalid code")
	}
	req, ok := p.dotabase.reqByID[id]
	if !ok {
		return nil, errors.New("Valid code but omg i lost the request")
	}
	return req, nil
}

// AuthRequestByID implements op.Storage.
func (p *provider) AuthRequestByID(ctx context.Context, authRequestID string) (op.AuthRequest, error) {
	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return nil, httputil.BadRequest("Invalid uuid")
	}
	p.dotabase.mu.Lock()
	req, ok := p.dotabase.reqByID[id]
	p.dotabase.mu.Unlock()
	if !ok {
		return nil, httputil.BadRequest("No request with that id")
	}
	return req, nil
}

// CreateAccessAndRefreshTokens implements op.Storage.
func (p *provider) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (accessTokenID string, newRefreshTokenID string, expiration time.Time, err error) {
	slog.Warn("oidcprovider.*service.CreateAccessAndRefreshTokens", "request", request, "currentRefreshToken", currentRefreshToken)
	panic("unimplemented")
}

// CreateAccessToken implements op.Storage.
func (p *provider) CreateAccessToken(ctx context.Context, request op.TokenRequest) (accessTokenID string, expiration time.Time, err error) {
	slog.Warn("oidcprovider.*service.CreateAccessToken", "request", request)
	return strings.Join(request.GetScopes(), " "), time.Now().Add(time.Minute), nil
}

// CreateAuthRequest implements op.Storage.
func (p *provider) CreateAuthRequest(ctx context.Context, r *oidc.AuthRequest, userID string) (op.AuthRequest, error) {
	if userID != "" {
		slog.Info("oidcprovider.*service.CreateAuthRequest: we got a userID!!!", "userID", userID)
	}

	id := uuid.New()
	req := authRequest{id: id, authCode: "", inner: r}
	p.dotabase.mu.Lock()
	p.dotabase.reqByID[id] = req
	p.dotabase.mu.Unlock()
	return req, nil
}

// DeleteAuthRequest implements op.Storage.
func (p *provider) DeleteAuthRequest(ctx context.Context, authRequestID string) error {
	slog.Warn("oidcprovider.*service.DeleteAuthRequest", "authRequestID", authRequestID)

	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	p.dotabase.mu.Lock()
	defer p.dotabase.mu.Unlock()
	req, ok := p.dotabase.reqByID[id]
	if !ok {
		return httputil.BadRequest("No request with that id")
	}
	delete(p.dotabase.reqByID, id)
	delete(p.dotabase.reqIdByAuthCode, req.authCode)
	return nil
}

// AuthorizeClientIDSecret implements op.Storage.
func (p *provider) AuthorizeClientIDSecret(ctx context.Context, clientID string, clientSecret string) error {
	id, err := base64.URLEncoding.DecodeString(clientID)
	if err != nil {
		return httputil.BadRequest("Invalid id format")
	}
	_, err = p.s.DB.GetClient(ctx, id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such client")
	}
	if err != nil {
		return err
	}
	secret, err := base64.URLEncoding.DecodeString(clientSecret)
	if err != nil {
		return httputil.BadRequest("Invalid secret format")
	}
	h := sha256.New()
	h.Write(secret)
	if subtle.ConstantTimeCompare(id, h.Sum(nil)) != 1 {
		return httputil.BadRequest("Invalid client secret")
	}
	return nil
}

// GetClientByClientID implements op.Storage.
func (p *provider) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
	id, err := base64.URLEncoding.DecodeString(clientID)
	if err != nil {
		return nil, httputil.BadRequest("Invalid id format")
	}
	c, err := p.s.DB.GetClient(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, httputil.BadRequest("No such client")
	}
	if err != nil {
		return nil, err
	}

	return client{
		id:           base64.URLEncoding.EncodeToString(c.ID),
		redirectURIs: c.RedirectUris,
	}, nil
}

// GetKeyByIDAndClientID implements op.Storage.
func (p *provider) GetKeyByIDAndClientID(ctx context.Context, keyID string, clientID string) (*jose.JSONWebKey, error) {
	slog.Warn("oidcprovider.*service.GetKeyByIDAndClientID", "keyID", keyID, "clientID", clientID)
	panic("unimplemented")
}

// GetPrivateClaimsFromScopes implements op.Storage.
func (p *provider) GetPrivateClaimsFromScopes(ctx context.Context, userID string, clientID string, scopes []string) (map[string]any, error) {
	slog.Warn("oidcprovider.*service.GetPrivateClaimsFromScopes", "userID", userID, "clientID", clientID, "scopes", scopes)
	panic("unimplemented")
}

// GetRefreshTokenInfo implements op.Storage.
func (p *provider) GetRefreshTokenInfo(ctx context.Context, clientID string, token string) (userID string, tokenID string, err error) {
	slog.Warn("oidcprovider.*service.GetRefreshTokenInfo", "clientID", clientID, "token", token)
	panic("unimplemented")
}

// Health implements op.Storage.
func (p *provider) Health(ctx context.Context) error {
	slog.Warn("oidcprovider.*service.Health")
	panic("unimplemented")
}

// RevokeToken implements op.Storage.
func (p *provider) RevokeToken(ctx context.Context, tokenOrTokenID string, userID string, clientID string) *oidc.Error {
	slog.Warn("oidcprovider.*service.RevokeToken", "tokenOrTokenID", tokenOrTokenID, "userID", userID, "clientID", clientID)
	panic("unimplemented")
}

// SaveAuthCode implements op.Storage.
func (p *provider) SaveAuthCode(ctx context.Context, authRequestID string, code string) error {
	slog.Warn("oidcprovider.*service.SaveAuthCode", "authRequestID", authRequestID, "code", code)
	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	p.dotabase.mu.Lock()
	defer p.dotabase.mu.Unlock()
	req, ok := p.dotabase.reqByID[id]
	if !ok {
		return httputil.BadRequest("No request with that id")
	}
	p.dotabase.reqIdByAuthCode[code] = id
	req.authCode = code
	p.dotabase.reqByID[id] = req
	return nil
}

// SetIntrospectionFromToken implements op.Storage.
func (p *provider) SetIntrospectionFromToken(ctx context.Context, userinfo *oidc.IntrospectionResponse, tokenID string, subject string, clientID string) error {
	slog.Warn("oidcprovider.*service.SetIntrospectionFromToken", "userinfo", userinfo, "tokenID", tokenID, "subject", subject, "clientID", clientID)
	panic("unimplemented")
}

// SetUserinfoFromScopes implements op.Storage.
func (p *provider) SetUserinfoFromScopes(ctx context.Context, userinfo *oidc.UserInfo, kthid string, clientID string, scopes []string) error {
	user, err := p.s.GetUser(ctx, kthid)
	if err != nil {
		return err
	}
	if user == nil {
		slog.Error("SetUserinfoFromScopes: user not found", "kthid", kthid, "clientID", clientID, "scopes", scopes)
		return errors.New("SetUserinfoFromScopes, no user but pretty sure that should have been handled in this request???")
	}
	for _, scope := range scopes {
		switch scope {
		case oidc.ScopeOpenID:
			userinfo.Subject = kthid
		case oidc.ScopeProfile:
			userinfo.GivenName = user.FirstName
			userinfo.FamilyName = user.FamilyName
		case oidc.ScopeEmail:
			userinfo.Email = user.Email
			userinfo.EmailVerified = true
		}
	}
	slog.Info("oidcprovider.*service.SetUserinfoFromScopes", "userinfo", userinfo)
	return nil
}

// SetUserinfoFromToken implements op.Storage.
func (p *provider) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID string, kthid string, origin string) error {
	slog.Warn("oidcprovider.*service.SetUserinfoFromToken", "userinfo", userinfo, "tokenID", tokenID, "subject", kthid, "origin", origin)
	user, err := p.s.GetUser(ctx, kthid)
	if err != nil {
		return err
	}
	if user == nil {
		slog.Error("SetUserinfoFromToken: user not found", "kthid", kthid)
		return errors.New("SetUserinfoFromToken, no user but pretty sure that should have been handled in this request???")
	}
	// TODO: Putting scopes in tokenID feels cursed
	scopes := strings.Split(tokenID, " ")
	for _, scope := range scopes {
		switch scope {
		case oidc.ScopeOpenID:
			userinfo.Subject = kthid
		case oidc.ScopeProfile:
			userinfo.GivenName = user.FirstName
			userinfo.FamilyName = user.FamilyName
		case oidc.ScopeEmail:
			userinfo.Email = user.Email
			userinfo.EmailVerified = true
		}
	}
	slog.Info("oidcprovider.*service.SetUserinfoFromToken", "userinfo", userinfo)
	return nil
}

// SignatureAlgorithms implements op.Storage.
func (p *provider) SignatureAlgorithms(ctx context.Context) ([]jose.SignatureAlgorithm, error) {
	return []jose.SignatureAlgorithm{jose.ES256}, nil
}

type publicKey struct{ *ecdsa.PublicKey }

func (s publicKey) ID() string                         { return "the-one-and-only" }
func (s publicKey) Key() any                           { return s.PublicKey }
func (s publicKey) Algorithm() jose.SignatureAlgorithm { return jose.ES256 }
func (s publicKey) Use() string                        { return "sig" }

var _ op.Key = publicKey{}

// KeySet implements op.Storage.
func (p *provider) KeySet(ctx context.Context) ([]op.Key, error) {
	return []op.Key{publicKey{&p.signingKey.PublicKey}}, nil
}

type privateKey struct{ *ecdsa.PrivateKey }

func (k privateKey) ID() string                                  { return "the-one-and-only" }
func (k privateKey) Key() any                                    { return k.PrivateKey }
func (k privateKey) SignatureAlgorithm() jose.SignatureAlgorithm { return jose.ES256 }

var _ op.SigningKey = privateKey{}

// SigningKey implements op.Storage.
func (p *provider) SigningKey(ctx context.Context) (op.SigningKey, error) {
	return privateKey{p.signingKey}, nil
}

// TerminateSession implements op.Storage.
func (p *provider) TerminateSession(ctx context.Context, userID string, clientID string) error {
	slog.Warn("oidcprovider.*service.TerminateSession", "userID", userID, "clientID", clientID)
	panic("unimplemented")
}

// TokenRequestByRefreshToken implements op.Storage.
func (p *provider) TokenRequestByRefreshToken(ctx context.Context, refreshTokenID string) (op.RefreshTokenRequest, error) {
	slog.Warn("oidcprovider.*service.TokenRequestByRefreshToken", "refreshTokenID", refreshTokenID)
	panic("unimplemented")
}

// ValidateJWTProfileScopes implements op.Storage.
func (p *provider) ValidateJWTProfileScopes(ctx context.Context, userID string, scopes []string) ([]string, error) {
	slog.Warn("oidcprovider.*service.ValidateJWTProfileScopes", "userID", userID, "scopes", scopes)
	panic("unimplemented")
}