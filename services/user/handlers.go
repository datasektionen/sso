package user

import (
	"net/http"
	"time"

	"github.com/datasektionen/logout/pkg/httputil"
)

func (s *service) index(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
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
		if hasCookie {
			http.SetCookie(w, &http.Cookie{Name: "return-path", MaxAge: -1})
		}
		http.Redirect(w, r, returnPath, http.StatusSeeOther)
		return nil
	}
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
	return index(s.passkey.LoginForm, s.dev.LoginForm)
}

func (s *service) account(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.Unauthorized()
	}
	passkeySettings, err := s.passkey.PasskeySettings(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return account(*user, passkeySettings)
}
