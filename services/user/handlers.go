package user

import (
	"log/slog"
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
)

func (s *service) index(r *http.Request) httputil.ToResponse {
	user, err := s.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	return Index(user)
}

func (s *service) logout(r *http.Request) httputil.ToResponse {
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

func (s *service) account(r *http.Request) httputil.ToResponse {
	user, err := s.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.Unauthorized()
	}
	passkeys, err := s.passkey.GetPasskeysForUser(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return Account(*user, passkeys)
}

func (s *service) showRegister(r *http.Request) httputil.ToResponse {
	return Register()
}

func (s *service) doRegister(r *http.Request) httputil.ToResponse {
	kthid := r.FormValue("kthid")
	if len(kthid) < 2 {
		return httputil.BadRequest("Invalid kthid")
	}
	if err := s.CreateUser(r.Context(), kthid); err != nil {
		return err
	}
	slog.Info("User registrated", "kthid", kthid)
	return http.RedirectHandler("/", http.StatusSeeOther)
}

func (s *service) showLoginDev(r *http.Request) httputil.ToResponse {
	return LoginDev()
}

func (s *service) doLoginDev(r *http.Request) httputil.ToResponse {
	user, err := s.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	sessionID, err := s.CreateSession(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: sessionID.String(),
			Path:  "/",
		})
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	})
}
