package dev

import (
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
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
		Name:  "session",
		Value: sessionID.String(),
		Path:  "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func (s *service) register(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	kthid := r.FormValue("kthid")
	if len(kthid) < 2 {
		return httputil.BadRequest("Invalid kthid")
	}
	if err := s.db.CreateUser(r.Context(), kthid); err != nil {
		return err
	}
	return http.RedirectHandler("/", http.StatusSeeOther)
}
