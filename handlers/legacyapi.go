package handlers

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/datasektionen/sso/pkg/config"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/pkg/pls"
	"github.com/datasektionen/sso/service"
	"github.com/google/uuid"
)

func legacyAPIHello(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return "Hello Login!!!!"
}

func legacyAPILogin(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
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
		Secure:   !config.Config.Dev,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	s.RedirectToLogin(w, r, "/legacyapi/callback")
	return nil
}

func legacyAPICallback(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	callback, _ := r.Cookie("legacyapi-callback")
	if callback == nil {
		return httputil.BadRequest("Idk where you came from. Maybe you took longer than 10 minutes?")
	}
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.BadRequest("You did not seem to get logged in")
	}
	token, err := s.DB.CreateToken(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return http.RedirectHandler(callback.Value+token.String(), http.StatusSeeOther)
}

func legacyAPIVerify(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	token, err := uuid.Parse(strings.TrimSuffix(r.PathValue("token"), ".json"))
	if err != nil {
		return httputil.BadRequest("Invalid token")
	}
	apiKey := r.FormValue("api_key")
	allowed, err := pls.CheckToken(r.Context(), apiKey, "login.login")
	if err != nil {
		return err
	}
	if !allowed {
		return httputil.Forbidden("API Key invalid or has incorrect permissions")
	}
	kthid, err := s.DB.GetToken(r.Context(), token)
	if err != nil {
		return err
	}
	user, err := s.GetUser(r.Context(), kthid)
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

func legacyAPILogout(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user != nil {
		if err := s.DB.DeleteToken(r.Context(), user.KTHID); err != nil {
			return err
		}
	}
	return s.Logout(w, r)
}
