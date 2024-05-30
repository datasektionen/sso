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

func BadRequest(message string) error {
	return HttpError{Message: message, StatusCode: http.StatusBadRequest}
}

func Unauthorized() error {
	return HttpError{StatusCode: http.StatusUnauthorized}
}
