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
	switch resp := resp.(type) {
	case templ.Component:
		if err := resp.Render(r.Context(), w); err != nil {
			Respond(err, w, r)
			return
		}
	case error:
		err := resp
		var httpErr HttpError
		if errors.As(err, &httpErr) {
			httpErr.ServeHTTP(w, r)
		} else {
			slog.Error("Error serving request", "path", r.URL.Path, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal server error"))
		}
	case string:
		s := resp
		_, _ = w.Write([]byte(s))
	case http.Handler:
		h := resp
		h.ServeHTTP(w, r)
	case jsonValue:
		j := resp.any
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(j); err != nil {
			slog.Error("Error writing json", "value", j)
		}
	case redirect:
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", resp.url)
		} else {
			http.Redirect(w, r, resp.url, http.StatusSeeOther)
		}
	default:
		slog.Error("Got invalid response type when serving request", "url", r.URL.String(), "response", resp)
	}
}

func Route[S interface {
	WithSession(r *http.Request) (*http.Request, error)
}](s S, f func(s S, w http.ResponseWriter, r *http.Request) ToResponse) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2, err := s.WithSession(r)
		if err != nil {
			Respond(err, w, r)
		}
		Respond(f(s, w, r2), w, r2)
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
	_, _ = w.Write([]byte(h.Message))
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

func NotFound() error {
	return HttpError{Message: "Not Found", StatusCode: http.StatusNotFound}
}

type redirect struct {
	url string
}

// A temporary redirect that will result in a full page redirect even when HTMX initiated the
// request
func Redirect(url string) ToResponse {
	return redirect{url}
}
