package legacyapi

import (
	"context"
	"crypto/rand"
	"net/http"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	user "github.com/datasektionen/logout/services/user/export"
)

type service struct {
	db      *database.Queries
	user    user.Service
	hmacKey [64]byte
}

func NewService(ctx context.Context, db *database.Queries) (*service, error) {
	// TODO: persist?
	var hmacKey [64]byte
	_, err := rand.Read(hmacKey[:])
	if err != nil {
		return nil, err
	}

	s := &service{db: db, hmacKey: hmacKey}

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
