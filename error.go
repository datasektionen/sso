package main

import "net/http"

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
	w.Write([]byte(h.Error()))
}

func BadRequest(message string) error {
	return HttpError{Message: message, StatusCode: http.StatusBadRequest}
}

func Unauthorized() error {
	return HttpError{StatusCode: http.StatusUnauthorized}
}

func Forbidden() error {
	return HttpError{StatusCode: http.StatusForbidden}
}
