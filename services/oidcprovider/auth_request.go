package oidcprovider

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

type authRequest struct {
	id       uuid.UUID
	authCode string
	inner    *oidc.AuthRequest
	kthid    string
}

var _ op.AuthRequest = authRequest{}

// Done implements op.AuthRequest.
func (a authRequest) Done() bool {
	slog.Warn("oidcprovider.authRequest.Done")
	return a.kthid != ""
}

// GetACR implements op.AuthRequest.
func (a authRequest) GetACR() string {
	return a.inner.ACRValues.String()
}

// GetAMR implements op.AuthRequest.
func (a authRequest) GetAMR() []string {
	return nil
}

// GetAudience implements op.AuthRequest.
func (a authRequest) GetAudience() []string {
	return []string{a.GetClientID()}
}

// GetAuthTime implements op.AuthRequest.
func (a authRequest) GetAuthTime() time.Time {
	// This makes the auth time not be set (it's not required by the standard).
	// Probably related to the fact that we set the clock scew to 10 seconds
	// but very cursed indeed.
	return time.Unix(10, 0)
}

// GetClientID implements op.AuthRequest.
func (a authRequest) GetClientID() string {
	return a.inner.ClientID
}

// GetCodeChallenge implements op.AuthRequest.
func (a authRequest) GetCodeChallenge() *oidc.CodeChallenge {
	slog.Warn("oidcprovider.authRequest.GetCodeChallenge")
	panic("unimplemented")
}

// GetID implements op.AuthRequest.
func (a authRequest) GetID() string {
	return a.id.String()
}

// GetNonce implements op.AuthRequest.
func (a authRequest) GetNonce() string {
	return a.inner.Nonce
}

// GetRedirectURI implements op.AuthRequest.
func (a authRequest) GetRedirectURI() string {
	return a.inner.RedirectURI
}

// GetResponseMode implements op.AuthRequest.
func (a authRequest) GetResponseMode() oidc.ResponseMode {
	return a.inner.ResponseMode
}

// GetResponseType implements op.AuthRequest.
func (a authRequest) GetResponseType() oidc.ResponseType {
	return a.inner.ResponseType
}

// GetScopes implements op.AuthRequest.
func (a authRequest) GetScopes() []string {
	return a.inner.Scopes
}

// GetState implements op.AuthRequest.
func (a authRequest) GetState() string {
	return a.inner.State
}

// GetSubject implements op.AuthRequest.
func (a authRequest) GetSubject() string {
	slog.Warn("oidcprovider.authRequest.GetSubject")
	return a.kthid
}
