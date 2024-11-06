package service

import (
	"context"
	"encoding/json"

	"github.com/a-h/templ"
	"github.com/datasektionen/logout/models"
	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/templates"
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
		if err := json.Unmarshal([]byte(passkey.Data), &c); err != nil {
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

func (s *Service) PasskeyLogin() func() templ.Component {
	return func() templ.Component { return templates.PasskeyLoginForm("", nil) }
}

func (s *Service) PasskeySettings(ctx context.Context, kthid string) (func() templ.Component, error) {
	passkeys, err := s.ListPasskeysForUser(ctx, kthid)
	if err != nil {
		return nil, err
	}
	return func() templ.Component { return templates.PasskeySettings(passkeys) }, nil
}
