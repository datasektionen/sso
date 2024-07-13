package user

import (
	"context"
	"net/http"
	"time"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	dev "github.com/datasektionen/logout/services/dev/export"
	"github.com/datasektionen/logout/services/user/export"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:generate templ generate

type service struct {
	db  *database.Queries
	dev dev.Service
}

var _ export.Service = &service{}

func NewService(db *database.Queries) (*service, error) {
	s := &service{db: db}

	http.Handle("GET /{$}", httputil.Route(s.index))
	http.Handle("GET /logout", httputil.Route(s.Logout))
	http.Handle("GET /account", httputil.Route(s.account))
	http.Handle("GET /invite/{id}", httputil.Route(s.acceptInvite))

	return s, nil
}

func (s *service) Assign(dev dev.Service) {
	s.dev = dev
}

func (s *service) GetUser(ctx context.Context, kthid string) (*export.User, error) {
	user, err := s.db.GetUser(ctx, kthid)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var memberTo time.Time
	if user.MemberTo.Valid {
		memberTo = user.MemberTo.Time
	}
	return &export.User{
		KTHID:      user.Kthid,
		UGKTHID:    user.UgKthid,
		Email:      user.Email,
		FirstName:  user.FirstName,
		FamilyName: user.FamilyName,
		YearTag:    user.YearTag,
		MemberTo:   memberTo,
		WebAuthnID: user.WebauthnID,
	}, nil
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
	return s.GetUser(r.Context(), kthid)
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
