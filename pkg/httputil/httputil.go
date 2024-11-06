package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

type ToResponse any

func Respond(resp ToResponse, w http.ResponseWriter, r *http.Request) {
	if resp == nil {
		return
	}
	switch resp.(type) {
	case templ.Component:
		resp.(templ.Component).Render(r.Context(), w)
	case error:
		err := resp.(error)
		var httpErr HttpError
		if errors.As(err, &httpErr) {
			httpErr.ServeHTTP(w, r)
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
	case jsonValue:
		j := resp.(jsonValue).any
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(j); err != nil {
			slog.Error("Error writing json", "value", j)
		}
	default:
		slog.Error("Got invalid response type when serving request", "url", r.URL.String(), "response", resp)
	}
}

func Route[S any](s S, f func(s S, w http.ResponseWriter, r *http.Request) ToResponse) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Respond(f(s, w, r), w, r)
	})
}

type jsonValue struct{ any }

func JSON(val any) jsonValue {
	return jsonValue{val}
}

type HttpError struct {
	Message    string
	StatusCode int
}

func (h HttpError) Error() string {
	s := http.StatusText(h.StatusCode)
	if h.Message != "" {
		s += ": " + h.Message
	}
	return s
}

func (h HttpError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(h.StatusCode)
	w.Write([]byte(h.Message))
}

func BadRequest(message string) error {
	return HttpError{Message: message, StatusCode: http.StatusBadRequest}
}

func Unauthorized() error {
	return HttpError{StatusCode: http.StatusUnauthorized}
}

func Forbidden(message string) error {
	return HttpError{Message: message, StatusCode: http.StatusForbidden}
}
