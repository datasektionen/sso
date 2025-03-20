package oidcprovider

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/models"
	"github.com/datasektionen/sso/pkg/config"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/pkg/pls"
	"github.com/datasektionen/sso/service"

	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
	"golang.org/x/text/language"
)

// http://localhost:7000/op/authorize?client_id=bing&response_type=token&scope=openid&redirect_uri=http://localhost:8080/callback

type provider struct {
	provider *op.Provider
	// dotabase dotabase
	rsaKey *rsa.PrivateKey
	s      *service.Service
}

// TODO: these should maybe not be stored in just a few hashmaps. One possible problem is that we never free them, which isn't optimal.
type dotabase struct {
	mu              sync.Mutex
	reqByID         map[uuid.UUID]authRequest
	reqIdByAuthCode map[string]uuid.UUID
	tokenByID       map[uuid.UUID]accessToken
}

type accessToken struct {
	kthid  string
	scopes []string
}

var _ op.Storage = &provider{}

func MountRoutes(s *service.Service) error {
	// Yes, the initialization of this key does indeed seem very shady. I do
	// however hope that if anything is done incorrectly, the
	// privateKey.Validate() should catch that. I didn't find a nice way to
	// initialize a private key from p and q and it is unneccesary to also
	// store d, so I guess I have to calculate it 🤷.
	// I guess one could also store a seed for a PRNG that is given to
	// `rsa.GenerateKey`, but I'm not completely sure about the security
	// implications of that.
	privateKey := rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: &big.Int{},
			E: 65537,
		},
		D:      &big.Int{},
		Primes: []*big.Int{},
	}
	{
		parts := strings.SplitN(config.Config.OIDCProviderKey, ",", 3)
		if len(parts) != 2 {
			return errors.New("Expected $OIDC_PROVIDER_KEY to have two comma-separated prime numbers in base 62")
		}
		var p, q big.Int
		p.SetString(parts[0], 62)
		q.SetString(parts[1], 62)
		privateKey.Primes = append(privateKey.Primes, &p, &q)
		privateKey.N.Mul(&p, &q)
		e := big.NewInt(int64(privateKey.E))
		pMinus1 := (&big.Int{}).Sub(&p, big.NewInt(1))
		qMinus1 := (&big.Int{}).Sub(&q, big.NewInt(1))
		phi := (&big.Int{}).Mul(pMinus1, qMinus1)
		privateKey.D.ModInverse(e, phi)
		if err := privateKey.Validate(); err != nil {
			return err
		}
		privateKey.Precompute()
	}
	if err := privateKey.Validate(); err != nil {
		return err
	}

	p := &provider{rsaKey: &privateKey, s: s}
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
			"pls_*",
		},
		SupportedScopes: []string{
			"openid", "profile", "email", "offline_access", "pls_*",
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

	kthid, err := p.s.GetLoggedInKTHID(r)
	if err != nil {
		return err
	}
	if kthid == "" {
		return httputil.BadRequest("User did not seem to get logged in")
	}

	if rows, err := p.s.DB.AuthRequestSetKTHID(r.Context(), database.AuthRequestSetKTHIDParams{
		ID:    id,
		Kthid: kthid,
	}); err != nil {
		return err
	} else if rows == 0 {
		return httputil.BadRequest("No request with that id")
	}

	http.Redirect(w, r, "/op"+op.AuthCallbackURL(p.provider)(r.Context(), authRequestID), http.StatusSeeOther)
	return nil
}

// AuthRequestByCode implements op.Storage.
func (p *provider) AuthRequestByCode(ctx context.Context, code string) (op.AuthRequest, error) {
	req, err := p.s.DB.GetAuthRequestByAuthCode(ctx, code)
	if err == pgx.ErrNoRows {
		return nil, httputil.BadRequest("Invalid code")
	} else if err != nil {
		return nil, err
	}
	return authRequestFromDB(req), nil
}

// AuthRequestByID implements op.Storage.
func (p *provider) AuthRequestByID(ctx context.Context, authRequestID string) (op.AuthRequest, error) {
	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return nil, httputil.BadRequest("Invalid uuid")
	}

	req, err := p.s.DB.GetAuthRequest(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, httputil.BadRequest("No request with that id")
	} else if err != nil {
		return nil, err
	}

	return authRequestFromDB(req), nil
}

// CreateAccessAndRefreshTokens implements op.Storage.
func (p *provider) CreateAccessAndRefreshTokens(ctx context.Context, request op.TokenRequest, currentRefreshToken string) (accessTokenID string, newRefreshTokenID string, expiration time.Time, err error) {
	slog.Warn("oidcprovider.*service.CreateAccessAndRefreshTokens", "request", request, "currentRefreshToken", currentRefreshToken)
	panic("unimplemented")
}

// CreateAccessToken implements op.Storage.
func (p *provider) CreateAccessToken(ctx context.Context, request op.TokenRequest) (accessTokenID string, expiration time.Time, err error) {
	slog.Warn("oidcprovider.*service.CreateAccessToken", "request", request)
	id, err := p.s.DB.CreateAccessToken(ctx, database.CreateAccessTokenParams{
		Kthid:  request.GetSubject(),
		Scopes: request.GetScopes(),
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return id.String(), time.Now().Add(time.Minute), nil
}

// CreateAuthRequest implements op.Storage.
func (p *provider) CreateAuthRequest(ctx context.Context, r *oidc.AuthRequest, userID string) (op.AuthRequest, error) {
	if userID != "" {
		slog.Info("oidcprovider.*service.CreateAuthRequest: we got a userID!!!", "userID", userID)
	}

	rr, _ := json.Marshal(r)
	req, err := p.s.DB.CreateAuthRequest(ctx, rr)
	if err != nil {
		return nil, err
	}
	return authRequestFromDB(req), nil
}

// DeleteAuthRequest implements op.Storage.
func (p *provider) DeleteAuthRequest(ctx context.Context, authRequestID string) error {
	slog.Warn("oidcprovider.*service.DeleteAuthRequest", "authRequestID", authRequestID)

	id, err := uuid.Parse(authRequestID)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
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
	client, err := p.s.DB.GetClient(ctx, clientID)
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
	if subtle.ConstantTimeCompare(client.SecretHash, h.Sum(nil)) != 1 {
		return httputil.BadRequest("Invalid client secret")
	}
	return nil
}

// GetClientByClientID implements op.Storage.
func (p *provider) GetClientByClientID(ctx context.Context, clientID string) (op.Client, error) {
	c, err := p.s.DB.GetClient(ctx, clientID)
	if err == pgx.ErrNoRows {
		return nil, httputil.BadRequest("No such client")
	}
	if err != nil {
		return nil, err
	}

	return client{
		id:           c.ID,
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

	if err := setUserinfo(ctx, userinfo, *user, scopes); err != nil {
		return err
	}
	slog.Info("oidcprovider.*service.SetUserinfoFromScopes", "userinfo", userinfo)
	return nil
}

// SetUserinfoFromToken implements op.Storage.
func (p *provider) SetUserinfoFromToken(ctx context.Context, userinfo *oidc.UserInfo, tokenID, kthid, origin string) error {
	slog.Warn("oidcprovider.*service.SetUserinfoFromToken", "tokenID", tokenID, "subject", kthid, "origin", origin)
	accessTokenID, err := uuid.Parse(tokenID)
	if err != nil {
		return httputil.BadRequest("SetUserinfoFromToken: invalid uuid syntax in token id")
	}
	p.dotabase.mu.Lock()
	token := p.dotabase.tokenByID[accessTokenID]
	p.dotabase.mu.Unlock()

	if token.kthid != kthid {
		return httputil.BadRequest("You're asking to get info about a different user than who the token is for")
	}

	user, err := p.s.GetUser(ctx, kthid)
	if err != nil {
		return err
	}
	if user == nil {
		slog.Error("SetUserinfoFromToken: user not found", "kthid", kthid)
		return errors.New("SetUserinfoFromToken, no user but pretty sure that should have been handled in this request???")
	}

	if err := setUserinfo(ctx, userinfo, *user, token.scopes); err != nil {
		return err
	}

	slog.Info("oidcprovider.*service.SetUserinfoFromToken", "userinfo", userinfo, "scopes", token.scopes)
	return nil
}

func setUserinfo(ctx context.Context, userinfo *oidc.UserInfo, user models.User, scopes []string) error {
	if userinfo.Claims == nil {
		userinfo.Claims = make(map[string]any)
	}

	for _, scope := range scopes {
		switch scope {
		case oidc.ScopeOpenID:
			userinfo.Subject = user.KTHID

		case oidc.ScopeProfile:
			userinfo.Name = user.FirstName + " " + user.FamilyName
			userinfo.GivenName = user.FirstName
			userinfo.FamilyName = user.FamilyName

		case oidc.ScopeEmail:
			userinfo.Email = user.Email
			userinfo.EmailVerified = true

		default:
			if group, ok := strings.CutPrefix(scope, "pls_"); ok {
				perms, err := pls.GetUserPermissionsForGroup(ctx, user.KTHID, group)
				if err != nil {
					slog.Error("setUserinfo: error getting permissions", "err", err)
					return err
				}
				userinfo.Claims[scope] = perms
			}
		}
	}

	return nil
}

// SignatureAlgorithms implements op.Storage.
func (p *provider) SignatureAlgorithms(ctx context.Context) ([]jose.SignatureAlgorithm, error) {
	return []jose.SignatureAlgorithm{jose.RS256}, nil
}

type publicKey struct{ privateKey }

func (s publicKey) ID() string                         { return "schmunguss-key" }
func (s publicKey) Key() any                           { return &s.PublicKey }
func (s publicKey) Algorithm() jose.SignatureAlgorithm { return jose.RS256 }
func (s publicKey) Use() string                        { return "sig" }

var _ op.Key = publicKey{}

// KeySet implements op.Storage.
func (p *provider) KeySet(ctx context.Context) ([]op.Key, error) {
	return []op.Key{publicKey{privateKey{p.rsaKey}}}, nil
}

type privateKey struct{ *rsa.PrivateKey }

func (k privateKey) ID() string                                  { return "schmunguss-key" }
func (k privateKey) Key() any                                    { return k.PrivateKey }
func (k privateKey) SignatureAlgorithm() jose.SignatureAlgorithm { return jose.RS256 }

var _ op.SigningKey = privateKey{}

// SigningKey implements op.Storage.
func (p *provider) SigningKey(ctx context.Context) (op.SigningKey, error) {
	return privateKey{p.rsaKey}, nil
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
