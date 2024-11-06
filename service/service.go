package service

import (
	"context"

	"github.com/datasektionen/logout/database"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
)

type Service struct {
	DB           *database.Queries
	WebAuthn     *webauthn.WebAuthn
	RelyingParty rp.RelyingParty
}

func NewService(ctx context.Context, db *database.Queries) (*Service, error) {
	wa, err := initWebAuthn()
	if err != nil {
		return nil, err
	}
	rp, err := initRP(ctx)
	if err != nil {
		return nil, err
	}
	return &Service{
		DB:           db,
		WebAuthn:     wa,
		RelyingParty: rp,
	}, nil
}
