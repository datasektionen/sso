package dev

import (
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/services/user/auth"
)

func (s *service) login(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.user.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	sessionID, err := s.db.CreateSession(r.Context(), user.KTHID)
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
