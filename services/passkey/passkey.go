package passkey

import (
	"net/http"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/services/passkey/export"
	user "github.com/datasektionen/logout/services/user/export"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jmoiron/sqlx"
)

//go:generate templ generate

type service struct {
	db       *sqlx.DB
	webauthn *webauthn.WebAuthn
	user     user.Service
}

var _ export.Service = &service{}

func NewService(db *sqlx.DB) (*service, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          config.Config.Origin.Hostname(),
		RPDisplayName: "Konglig Datasektionen",
		RPOrigins:     []string{config.Config.Origin.String()},
	})
	if err != nil {
		return nil, err
	}

	s := &service{db: db, webauthn: wa}

	if err := s.migrateDB(); err != nil {
		return nil, err
	}

	http.Handle("GET /login/passkey", httputil.Route(s.beginLoginPasskey))
	http.Handle("POST /login/passkey", httputil.Route(s.finishLoginPasskey))
	http.Handle("GET /passkey/add", httputil.Route(s.beginAddPasskey))
	http.Handle("POST /passkey/add", httputil.Route(s.finishAddPasskey))
	http.Handle("POST /passkey/remove", httputil.Route(s.removePasskey))

	return s, nil
}

func (s *service) Assign(user user.Service) {
	s.user = user
}
