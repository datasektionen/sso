package user

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/datasektionen/logout/pkg/httputil"
)

func (s *service) index(r *http.Request) httputil.ToResponse {
	returnPath := r.FormValue("return-path")
	if returnPath != "" && returnPath[0] != '/' {
		return httputil.BadRequest("Invalid return path")
	}
	hasCookie := false
	if returnPath == "" {
		c, _ := r.Cookie("return-path")
		if c != nil {
			returnPath = c.Value
			hasCookie = true
		}
	}
	if returnPath == "" {
		returnPath = "/account"
	}
	if kthid, err := s.GetLoggedInKTHID(r); err != nil {
		return err
	} else if kthid != "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hasCookie {
				http.SetCookie(w, &http.Cookie{Name: "return-path", MaxAge: -1})
			}
			http.Redirect(w, r, returnPath, http.StatusSeeOther)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if returnPath != "" {
			http.SetCookie(w, &http.Cookie{
				Name:     "return-path",
				Value:    returnPath,
				MaxAge:   int((time.Minute * 10).Seconds()),
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}
		httputil.Route(func(r *http.Request) httputil.ToResponse {
			user, err := s.GetLoggedInUser(r)
			if err != nil {
				return err
			}
			return Index(user)
		}).ServeHTTP(w, r)
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
	if err := s.createUser(r.Context(), kthid); err != nil {
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
	sessionID, err := s.createSession(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: sessionID.String(),
			Path:  "/",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
