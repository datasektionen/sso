package user

import (
	"net/http"
	"time"

	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/pkg/pls"
	"github.com/datasektionen/logout/service"
	"github.com/datasektionen/logout/templates"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func MountRoutes(s *service.Service) {
	http.Handle("GET /{$}", httputil.Route(s, index))
	http.Handle("GET /logout", httputil.Route(s, logout))
	http.Handle("GET /account", httputil.Route(s, account))
	http.Handle("GET /invite/{id}", httputil.Route(s, acceptInvite))
}

const nextUrlCookie string = "_logout_next-url"

func validNextURL(url string) bool {
	if url == "" {
		return true
	}
	// The "url" must be a path (and possibly params)
	if len(url) > 0 && url[0] != '/' {
		return false
	}
	// If it starts with two slashes there will be an implicit `https:` in front, so then it's not a path
	if len(url) > 1 && url[1] == '/' {
		return false
	}
	return true
}

func index(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	nextURL := r.FormValue("next-url")
	if !validNextURL(nextURL) {
		return httputil.BadRequest("Invalid return url")
	}
	hasCookie := false
	if nextURL == "" {
		c, _ := r.Cookie(nextUrlCookie)
		if c != nil {
			nextURL = c.Value
			hasCookie = true
		}
	}
	if nextURL == "" {
		nextURL = "/account"
	}
	if kthid, err := s.GetLoggedInKTHID(r); err != nil {
		return err
	} else if kthid != "" {
		if hasCookie {
			http.SetCookie(w, &http.Cookie{Name: nextUrlCookie, MaxAge: -1})
		}
		http.Redirect(w, r, nextURL, http.StatusSeeOther)
		return nil
	}
	if nextURL != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     nextUrlCookie,
			Value:    nextURL,
			MaxAge:   int((time.Minute * 10).Seconds()),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	return templates.Index(s.PasskeyLogin(), s.DevLoginForm)
}

func logout(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return s.Logout(w, r)
}

func account(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	passkeySettings, err := s.PasskeySettings(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	isAdmin, err := pls.CheckUser(r.Context(), user.KTHID, "admin-read")
	return templates.Account(*user, passkeySettings, isAdmin)
}

func acceptInvite(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	idString := r.PathValue("id")
	if idString == "-" {
		idCookie, _ := r.Cookie("invite")
		if idCookie == nil {
			return httputil.BadRequest("No invite id found")
		}
		idString = idCookie.Value
	}
	id, err := uuid.Parse(idString)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	inv, err := s.DB.GetInvite(r.Context(), id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such invite")
	} else if err != nil {
		return err
	}
	if time.Now().After(inv.ExpiresAt.Time) {
		return httputil.BadRequest("Invite expired")
	}
	if inv.MaxUses.Valid && inv.CurrentUses >= inv.MaxUses.Int32 {
		return httputil.BadRequest("This invite cannot be used to create more users")
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "invite",
		Value:    id.String(),
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	return templates.AcceptInvite()
}
