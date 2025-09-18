package service

import (
	"context"
	"encoding/json"

	"github.com/datasektionen/sso/models"
	"github.com/datasektionen/sso/pkg/config"
	"github.com/go-webauthn/webauthn/webauthn"
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
			ID:   passkey.ID,
			Name: passkey.Name,
			Cred: c,
		}
	}
	return passkeys, nil
}
