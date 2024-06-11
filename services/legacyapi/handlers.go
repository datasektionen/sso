package legacyapi

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/google/uuid"
)

func (s *service) hello(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return "Hello Login!!!!"
}

func (s *service) login(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	callbackString := r.FormValue("callback")
	callback, err := url.Parse(callbackString)
	if err != nil {
		return httputil.BadRequest("Invalid callback url")
	}
	if callback.Scheme == "http" && callback.Hostname() != "localhost" && callback.Hostname() != "127.0.0.1" {
		callback.Scheme = "https"
	}
	if !slices.Contains([]string{"http", "https"}, callback.Scheme) {
		return httputil.BadRequest("Invalid callback url")
	}
	if !(slices.ContainsFunc([]string{
		"datasektionen.se",
		"dsekt.se",
		"ddagen.se",
		"d-dagen.se",
		"metaspexet.se",
		"betaspexet.se",
		"betasektionen.se",
	}, func(domainName string) bool { return strings.HasSuffix("."+callback.Hostname(), "."+domainName) }) ||
		callback.Hostname() == "localhost" ||
		callback.Hostname() == "127.0.0.1") {
		return httputil.BadRequest("Invalid callback url")
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "legacyapi-callback",
		Value:    callback.String(),
		Path:     "/legacyapi",
		MaxAge:   int((time.Minute * 10).Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/?return-path="+url.QueryEscape("/legacyapi/callback"), http.StatusSeeOther)
	return nil
}

func (s *service) callback(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	callback, _ := r.Cookie("legacyapi-callback")
	if callback == nil {
		return httputil.BadRequest("Idk where you came from. Maybe you took longer than 10 minutes?")
	}
	user, err := s.user.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("You did not seem to get logged in")
	}
	token, err := s.db.CreateToken(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return http.RedirectHandler(callback.Value+token.String(), http.StatusSeeOther)
}

func (s *service) verify(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	token, err := uuid.Parse(strings.TrimSuffix(r.PathValue("token"), ".json"))
	if err != nil {
		return httputil.BadRequest("Invalid token")
	}
	apiKey := r.FormValue("api_key")
	_ = apiKey // TODO: verify dis boi
	kthid, err := s.db.GetToken(r.Context(), token)
	if err != nil {
		return err
	}
	user, err := s.user.GetUser(r.Context(), kthid)
	if err != nil {
		return err
	}
	return httputil.JSON(map[string]any{
		"first_name": user.FirstName,
		"last_name":  user.FamilyName,
		"user":       kthid,
		"emails":     user.Email,
		"ugkthid":    user.UGKTHID,
	})
}

func (s *service) logout(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.user.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user != nil {
		if err := s.db.DeleteToken(r.Context(), user.KTHID); err != nil {
			return err
		}
	}
	return s.user.Logout(w, r)
}
