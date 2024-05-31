package user

import (
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
	passkey "github.com/datasektionen/logout/services/passkey/export"
	"github.com/datasektionen/logout/services/user/export"
	"github.com/jmoiron/sqlx"
)

//go:generate templ generate

type service struct {
	db      *sqlx.DB
	passkey passkey.Service
}

var _ export.Service = &service{}

func NewService(db *sqlx.DB) (*service, error) {
	s := &service{db: db}

	if err := s.migrateDB(); err != nil {
		return nil, err
	}

	http.Handle("GET /{$}", httputil.Route(s.index))
	http.Handle("GET /logout", httputil.Route(s.Logout))
	http.Handle("GET /account", httputil.Route(s.account))
	http.Handle("GET /register", httputil.Route(s.showRegister))
	http.Handle("POST /register", httputil.Route(s.doRegister))
	http.Handle("GET /login/dev", httputil.Route(s.showLoginDev))
	http.Handle("POST /login/dev", httputil.Route(s.doLoginDev))

	return s, nil
}

func (s *service) Assign(passkey passkey.Service) {
	s.passkey = passkey
}

func (s *service) GetLoggedInKTHID(r *http.Request) (string, error) {
	sessionCookie, _ := r.Cookie("session")
	if sessionCookie == nil {
		return "", nil
	}
	return s.GetSession(sessionCookie.Value)
}

func (s *service) GetLoggedInUser(r *http.Request) (*export.User, error) {
	kthid, err := s.GetLoggedInKTHID(r)
	if err != nil {
		return nil, err
	}
	user, err := s.GetUser(r.Context(), kthid)
	return user, nil
}

func (s *service) Logout(r *http.Request) httputil.ToResponse {
	sessionCookie, _ := r.Cookie("session")
	if sessionCookie != nil {
		if err := s.RemoveSession(r.Context(), sessionCookie.Value); err != nil {
			return err
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

