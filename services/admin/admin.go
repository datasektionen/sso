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

	http.Handle("GET /admin/members", s.auth(httputil.Route(s.members)))
	http.Handle("POST /admin/members/upload-sheet", s.auth(httputil.Route(s.uploadSheet)))
	http.Handle("GET /admin/members/upload-sheet", s.auth(httputil.Route(s.processSheet)))

	http.Handle("GET /admin/oidc-clients", s.auth(httputil.Route(s.oidcClients)))
	http.Handle("GET /admin/list-oidc-clients", s.auth(httputil.Route(s.listOIDCClients)))
	http.Handle("POST /admin/oidc-clients", s.auth(httputil.Route(s.createOIDCClient)))
	http.Handle("DELETE /admin/oidc-clients/{id}", s.auth(httputil.Route(s.deleteOIDCClient)))

	http.Handle("GET /admin/invites", s.auth(httputil.Route(s.invites)))
	http.Handle("GET /admin/invites/{id}", s.auth(httputil.Route(s.invite)))
	http.Handle("POST /admin/invites", s.auth(httputil.Route(s.createInvite)))
	http.Handle("DELETE /admin/invites/{id}", s.auth(httputil.Route(s.deleteInvite)))
	http.Handle("GET /admin/invites/{id}/edit", s.auth(httputil.Route(s.editInviteForm)))
	http.Handle("PUT /admin/invites/{id}", s.auth(httputil.Route(s.updateInvite)))

	return s, nil
}

func (s *service) Assign(user user.Service) {
	s.user = user
}
