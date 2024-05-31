package legacyapi

import (
	"context"
	"crypto/rand"
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
	user "github.com/datasektionen/logout/services/user/export"
	"github.com/jmoiron/sqlx"
)

type service struct {
	db      *sqlx.DB
	user    user.Service
	hmacKey [64]byte
}

func NewService(ctx context.Context, db *sqlx.DB) (*service, error) {
	// TODO: persist?
	var hmacKey [64]byte
	_, err := rand.Read(hmacKey[:])
	if err != nil {
		return nil, err
	}

	s := &service{db: db, hmacKey: hmacKey}

	if err := s.migrateDB(ctx); err != nil {
		return nil, err
	}

	http.Handle("/legacyapi/hello", httputil.Route(s.hello))
	http.Handle("/legacyapi/login", httputil.Route(s.login))
	http.Handle("/legacyapi/callback", httputil.Route(s.callback))
	http.Handle("/legacyapi/verify/{token}", httputil.Route(s.verify))
	http.Handle("/legacyapi/logout", httputil.Route(s.logout))

	return s, nil
}

func (s *service) Assign(user user.Service) {
	s.user = user
}
