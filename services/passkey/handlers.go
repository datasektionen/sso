package passkey

import (
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/services/passkey/export"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

var hackSession *webauthn.SessionData

func (s *service) beginLoginPasskey(r *http.Request) httputil.ToResponse {
	user, err := s.user.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	passkeys, err := s.GetPasskeysForUser(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	if len(passkeys) == 0 {
		return httputil.BadRequest("You have no registered passkeys")
	}
	credAss, sessionData, err := s.webauthn.BeginLogin(export.WebAuthnUser{User: user, Passkeys: passkeys})
	hackSession = sessionData
	if err != nil {
		return err
	}
	return LoginPasskey(credAss)
}

func (s *service) finishLoginPasskey(r *http.Request) httputil.ToResponse {
	user, err := s.user.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	passkeys, err := s.GetPasskeysForUser(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	_, err = s.webauthn.FinishLogin(export.WebAuthnUser{User: user, Passkeys: passkeys}, *hackSession, r)
	if err != nil {
		return err
	}
	return s.user.LoginUser(r.Context(), user.KTHID)
}

func (s *service) beginAddPasskey(r *http.Request) httputil.ToResponse {
	user, err := s.user.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.Unauthorized()
	}
	creation, sessionData, err := s.webauthn.BeginRegistration(export.WebAuthnUser{User: user})
	if err != nil {
		return err
	}
	hackSession = sessionData

	return AddPasskey(creation)
}

func (s *service) finishAddPasskey(r *http.Request) httputil.ToResponse {
	user, err := s.user.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.Unauthorized()
	}
	// passkeys aren't gotten from within this function
	cred, err := s.webauthn.FinishRegistration(export.WebAuthnUser{User: user}, *hackSession, r)
	if err != nil {
		return err
	}
	name := r.FormValue("name")
	if err := s.AddPasskey(r.Context(), user.KTHID, name, *cred); err != nil {
		return err
	}
	return ""
}

func (s *service) removePasskey(r *http.Request) httputil.ToResponse {
	user, err := s.user.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.Unauthorized()
	}
	passkeyID, err := uuid.Parse(r.FormValue("passkey-id"))
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	if err := s.RemovePasskey(r.Context(), user.KTHID, passkeyID); err != nil {
		return err
	}
	return http.RedirectHandler("/account", http.StatusSeeOther)
}
