package export

import (
	"context"

	"github.com/a-h/templ"
	user "github.com/datasektionen/logout/services/user/export"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type Service interface {
	PasskeyLogin() func() templ.Component
	PasskeySettings(ctx context.Context, kthid string) (func() templ.Component, error)
}

type Passkey struct {
	ID   uuid.UUID           `json:"id"`
	Name string              `json:"name"`
	Cred webauthn.Credential `json:"-"`
}

type WebAuthnUser struct {
	User     *user.User
	Passkeys []Passkey
}

func (u WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	res := make([]webauthn.Credential, len(u.Passkeys))
	for i, cred := range u.Passkeys {
		res[i] = cred.Cred
	}
	return res
}

func (u WebAuthnUser) WebAuthnDisplayName() string {
	return u.User.FirstName + " " + u.User.FamilyName
}

func (u WebAuthnUser) WebAuthnID() []byte {
	return u.User.WebAuthnID[:]
}

func (u WebAuthnUser) WebAuthnIcon() string {
	// NOTE: No loger in the spec, library recommends empty string
	return ""
}

func (u WebAuthnUser) WebAuthnName() string {
	return u.User.KTHID
}
