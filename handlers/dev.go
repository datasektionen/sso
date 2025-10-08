package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
)

func devLogin(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	return s.LoginUser(r.Context(), user.KTHID, true)
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
