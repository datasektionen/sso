package service

import (
	"context"
	"encoding/json"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/models"
	"github.com/datasektionen/sso/pkg/config"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
)

func initWebAuthn() (*webauthn.WebAuthn, error) {
	return webauthn.New(&webauthn.Config{
		RPID:          config.Config.Origin.Hostname(),
		RPDisplayName: "Konglig Datasektionen",
		RPOrigins:     []string{config.Config.Origin.String()},
	})
}

func (s *Service) ListPasskeysForUser(ctx context.Context, kthid string) ([]models.Passkey, error) {
	dbPasskeys, err := s.DB.ListPasskeysByUser(ctx, kthid)
	if err != nil {
		return nil, err
	}
	passkeys := make([]models.Passkey, len(dbPasskeys))
	for i, passkey := range dbPasskeys {
		var c webauthn.Credential
		if err := json.Unmarshal(passkey.Data, &c); err != nil {
			return nil, err
		}
		passkeys[i] = models.Passkey{
			ID:           passkey.ID,
			Name:         passkey.Name,
			Cred:         c,
			Discoverable: passkey.Discoverable,
		}
	}
	return passkeys, nil
}

func (s *Service) GetPasskey(ctx context.Context, kthid string, rawID []byte) (*models.Passkey, error) {
	passkey, err := s.DB.GetPasskey(ctx, database.GetPasskeyParams{
		ID:    rawID,
		Kthid: kthid,
	})
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var c webauthn.Credential
	if err := json.Unmarshal(passkey.Data, &c); err != nil {
		return nil, err
	}
	return &models.Passkey{
		ID:           passkey.ID,
		Name:         passkey.Name,
		Cred:         c,
		Discoverable: passkey.Discoverable,
	}, nil
}
