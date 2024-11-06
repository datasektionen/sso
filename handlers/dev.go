package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/datasektionen/logout/pkg/auth"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/service"
)

func devLogin(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
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

func autoReload(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	flusher, canFlush := w.(interface{ Flush() })
	if !canFlush {
		return errors.New("Cannot flush body")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	for {
		_, _ = w.Write([]byte("data: schmunguss\n\n"))
		if canFlush {
			flusher.Flush()
		}
		time.Sleep(time.Minute)
	}
}
