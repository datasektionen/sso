package auth

import "net/http"

const SessionCookieName string = "_sso_session"

func SessionCookie(sessionID string) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
