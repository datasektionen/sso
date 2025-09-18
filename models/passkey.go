package models

import (
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type Passkey struct {
	ID           uuid.UUID           `json:"id"`
	Name         string              `json:"name"`
	Cred         webauthn.Credential `json:"-"`
	Discoverable bool                `json:"discoverable"`
}

type WebAuthnUser struct {
	User     *User
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
