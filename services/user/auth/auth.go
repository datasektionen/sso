package auth

import "net/http"

const SESSION_COOKIE string = "_logout_session"

func UserCookie(sessionID string) *http.Cookie {
	return &http.Cookie{
		Name:     SESSION_COOKIE,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
