package dev

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"
	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/services/dev/export"
	user "github.com/datasektionen/logout/services/user/export"
	"github.com/datasektionen/logout/templates"
)

//go:generate templ generate

type service struct {
	db   *database.Queries
	user user.Service
}

var _ export.Service = &service{}

func NewService(db *database.Queries) (*service, error) {
	s := &service{db: db}

	if config.Config.Dev {
		http.Handle("POST /login/dev", httputil.Route(s.login))
	}

	return s, nil
}

func (s *service) Assign(user user.Service) {
	s.user = user
}

func (s *service) LoginForm() templ.Component {
	if !config.Config.Dev {
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			return nil
		})
	}
	return templates.DevLoginForm()
}
