package auth

import (
	"net/http"

	"github.com/datasektionen/sso/pkg/config"
)

const SessionCookieName string = "_sso_session"

func SessionCookie(sessionID string) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   !config.Config.Dev,
		SameSite: http.SameSiteLaxMode,
	}
}
