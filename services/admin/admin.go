package admin

import (
	"net/http"
	"sync"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
)

//go:generate templ generate

type service struct {
	db          *database.Queries
	memberSheet memberSheet
}

type memberSheet struct {
	mu         sync.Mutex
	data       []byte
	inProgress bool
}

func NewService(db *database.Queries) (*service, error) {
	s := &service{db: db}

	http.Handle("GET /admin", httputil.Route(s.admin))
	http.Handle("POST /admin/upload-sheet", httputil.Route(s.uploadSheet))
	http.Handle("GET /admin/upload-sheet", httputil.Route(s.processSheet))

	return s, nil
}

func (s *service) Assign() {}
