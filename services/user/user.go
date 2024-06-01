package user

import (
	"context"
	"net/http"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	passkey "github.com/datasektionen/logout/services/passkey/export"
	"github.com/datasektionen/logout/services/user/export"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:generate templ generate

type service struct {
	db      *database.Queries
	passkey passkey.Service
}

var _ export.Service = &service{}

func NewService(db *database.Queries) (*service, error) {
	s := &service{db: db}

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

func (s *service) GetUser(ctx context.Context, kthid string) (*export.User, error) {
	user, err := s.db.GetUser(ctx, kthid)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &export.User{KTHID: user.Kthid, WebAuthnID: user.WebauthnID}, nil
}

func (s *service) LoginUser(ctx context.Context, kthid string) httputil.ToResponse {
	sessionID, err := s.db.CreateSession(ctx, kthid)
	if err != nil {
		return err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionID.String(),
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func (s *service) GetLoggedInKTHID(r *http.Request) (string, error) {
	sessionCookie, _ := r.Cookie("session")
	if sessionCookie == nil {
		return "", nil
	}
	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		return "", nil
	}
	session, err := s.db.GetSession(r.Context(), sessionID)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return session, nil
}

func (s *service) GetLoggedInUser(r *http.Request) (*export.User, error) {
	kthid, err := s.GetLoggedInKTHID(r)
	if err != nil {
		return nil, err
	}
	if kthid == "" {
		return nil, nil
	}
	user, err := s.db.GetUser(r.Context(), kthid)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &export.User{KTHID: user.Kthid, WebAuthnID: user.WebauthnID}, nil
}

func (s *service) Logout(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	sessionCookie, _ := r.Cookie("session")
	if sessionCookie != nil {
		sessionID, err := uuid.Parse(sessionCookie.Value)
		if err != nil {
			if err := s.db.RemoveSession(r.Context(), sessionID); err != nil {
				return err
			}
		}
	}
	http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
