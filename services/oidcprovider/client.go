package oidcprovider

import (
	"log/slog"
	"net/url"
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

type client struct {
	id string
}

var _ op.Client = client{}

// AccessTokenType implements op.Client.
func (c client) AccessTokenType() op.AccessTokenType {
	return op.AccessTokenTypeBearer
}

// ApplicationType implements op.Client.
func (c client) ApplicationType() op.ApplicationType {
	slog.Warn("oidcprovider.client.ApplicationType")
	return op.ApplicationTypeWeb
}

// AuthMethod implements op.Client.
func (c client) AuthMethod() oidc.AuthMethod {
	return oidc.AuthMethodBasic
}

// ClockSkew implements op.Client.
func (c client) ClockSkew() time.Duration {
	return time.Second * 10
}

// DevMode implements op.Client.
func (c client) DevMode() bool {
	slog.Warn("oidcprovider.client.DevMode")
	return true
}

// GetID implements op.Client.
func (c client) GetID() string {
	return c.id
}

// GrantTypes implements op.Client.
func (c client) GrantTypes() []oidc.GrantType {
	slog.Warn("oidcprovider.client.GrantTypes")
	return []oidc.GrantType{oidc.GrantTypeCode}
}

// IDTokenLifetime implements op.Client.
func (c client) IDTokenLifetime() time.Duration {
	slog.Warn("oidcprovider.client.IDTokenLifetime")
	return 0
}

// IDTokenUserinfoClaimsAssertion implements op.Client.
func (c client) IDTokenUserinfoClaimsAssertion() bool {
	return true
}

// IsScopeAllowed implements op.Client.
func (c client) IsScopeAllowed(scope string) bool {
	slog.Warn("oidcprovider.client.IsScopeAllowed", "scope", scope)
	panic("unimplemented")
}

// LoginURL implements op.Client.
func (c client) LoginURL(authRequestID string) string {
	return "/?return-path=" + url.QueryEscape("/op-callback?auth-request-id="+authRequestID)
}

// PostLogoutRedirectURIs implements op.Client.
func (c client) PostLogoutRedirectURIs() []string {
	slog.Warn("oidcprovider.client.PostLogoutRedirectURIs")
	panic("unimplemented")
}

// RedirectURIs implements op.Client.
func (c client) RedirectURIs() []string {
	slog.Warn("oidcprovider.client.RedirectURIs")
	return []string{"http://localhost:3000/oidc/callback"}
}

// ResponseTypes implements op.Client.
func (c client) ResponseTypes() []oidc.ResponseType {
	return []oidc.ResponseType{"code"}
}

// RestrictAdditionalAccessTokenScopes implements op.Client.
func (c client) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	slog.Warn("oidcprovider.client.RestrictAdditionalAccessTokenScopes")
	panic("unimplemented")
}

// RestrictAdditionalIdTokenScopes implements op.Client.
func (c client) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	slog.Warn("oidcprovider.client.RestrictAdditionalIdTokenScopes")
	return func(scopes []string) []string { return scopes }
}
