package admin

import (
	"net/http"
	"sync"

	"github.com/a-h/templ"
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
	// This is locked when retrieving or assigning the reader channel.
	mu sync.Mutex
	// The post handler will instantiate this channel and finish the response.
	// It will then wait for an "event channel" to be sent on this channel
	// until it begins processing the uploaded sheet and then continuously send
	// events from the sheet handling on the "event channel". After processing,
	// this channel will be closed and reassigned to nil.
	//
	// The get handler will send an "event channel" on this channel, read
	// events from that and send them along to the client with SSE.
	reader chan<- chan<- sheetEvent
}

type sheetEvent struct {
	name string
	component templ.Component
}

func NewService(db *database.Queries) (*service, error) {
	s := &service{db: db}

	http.Handle("GET /admin", s.auth(httputil.Route(s.admin)))

	http.Handle("GET /admin/members", s.auth(httputil.Route(s.members)))
	http.Handle("POST /admin/members/upload-sheet", s.auth(httputil.Route(s.uploadSheet)))
	http.Handle("GET /admin/members/upload-sheet", s.auth(httputil.Route(s.processSheet)))

	http.Handle("GET /admin/oidc-clients", s.auth(httputil.Route(s.oidcClients)))
	http.Handle("POST /admin/oidc-clients", s.auth(httputil.Route(s.createOIDCClient)))
	http.Handle("DELETE /admin/oidc-clients/{id}", s.auth(httputil.Route(s.deleteOIDCClient)))
	http.Handle("POST /admin/oidc-clients/{id}", s.auth(httputil.Route(s.addRedirectURI)))
	http.Handle("DELETE /admin/oidc-clients/{id}/{uri}", s.auth(httputil.Route(s.removeRedirectURI)))

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
