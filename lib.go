package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

type ToResponse any

func route(f func(r *http.Request) ToResponse) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := f(r)
		switch resp.(type) {
		case templ.Component:
			resp.(templ.Component).Render(r.Context(), w)
		case error:
			err := resp.(error)
			var httpErr HttpError
			if errors.As(err, &httpErr) {
				w.WriteHeader(httpErr.StatusCode)
				w.Write([]byte(httpErr.Error()))
			} else {
				slog.Error("Error serving request", "path", r.URL.Path, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal server error"))
			}
		case string:
			s := resp.(string)
			w.Write([]byte(s))
		case http.Handler:
			h := resp.(http.Handler)
			h.ServeHTTP(w, r)
		}
	})
}
