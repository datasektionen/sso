package dev

import (
	"net/http"

	"github.com/datasektionen/logout/pkg/auth"
	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/service"
)

func MountRoutes(s *service.Service) {
	if config.Config.Dev {
		http.Handle("POST /login/dev", httputil.Route(s, login))
	}
}

func login(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	sessionID, err := s.DB.CreateSession(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:  auth.SessionCookieName,
		Value: sessionID.String(),
		Path:  "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
