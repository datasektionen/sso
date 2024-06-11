package admin

import (
	"net/http"
	"sync"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	user "github.com/datasektionen/logout/services/user/export"
)

//go:generate templ generate

type service struct {
	db          *database.Queries
	memberSheet memberSheet
	user        user.Service
}

type memberSheet struct {
	mu         sync.Mutex
	data       []byte
	inProgress bool
}

func NewService(db *database.Queries) (*service, error) {
	s := &service{db: db}

	http.Handle("GET /admin", s.auth(httputil.Route(s.admin)))
	http.Handle("POST /admin/upload-sheet", s.auth(httputil.Route(s.uploadSheet)))
	http.Handle("GET /admin/upload-sheet", s.auth(httputil.Route(s.processSheet)))

	http.Handle("GET /admin/oidc-clients", s.auth(httputil.Route(s.oidcClients)))
	http.Handle("POST /admin/oidc-clients", s.auth(httputil.Route(s.createOIDCClient)))
	http.Handle("DELETE /admin/oidc-clients/{id}", s.auth(httputil.Route(s.deleteOIDCClient)))

	return s, nil
}

func (s *service) Assign(user user.Service) {
	s.user = user
}
