package passkey

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/services/passkey/export"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

var hackSession *webauthn.SessionData

func (s *service) beginLoginPasskey(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.user.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	passkeys, err := s.listPasskeysForUser(r.Context(), user.KTHID)
	if len(passkeys) == 0 {
		return httputil.BadRequest("You have no registered passkeys")
	}
	credAss, sessionData, err := s.webauthn.BeginLogin(export.WebAuthnUser{User: user, Passkeys: passkeys})
	hackSession = sessionData
	if err != nil {
		return err
	}
	return httputil.JSON(credAss)
}

func (s *service) finishLoginPasskey(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	var body struct {
		KTHID string `json:"kthid"`
		protocol.CredentialAssertionResponse
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return httputil.BadRequest("Invalid credential")
	}
	credAss, err := body.CredentialAssertionResponse.Parse()
	if err != nil {
		return httputil.BadRequest("Invalid credential")
	}

	user, err := s.user.GetUser(r.Context(), body.KTHID)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	passkeys, err := s.listPasskeysForUser(r.Context(), user.KTHID)
	if err != nil {
		return err
	}

	_, err = s.webauthn.ValidateLogin(export.WebAuthnUser{User: user, Passkeys: passkeys}, *hackSession, credAss)
	if err != nil {
		return err
	}
	return s.user.LoginUser(r.Context(), user.KTHID)
}

func (s *service) beginAddPasskey(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
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

	return httputil.JSON(creation)
}

func (s *service) finishAddPasskey(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.user.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.Unauthorized()
	}

	var body struct {
		Name string `json:"name"`
		protocol.CredentialCreationResponse
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return httputil.BadRequest("Invalid credential")
	}
	cc, err := body.CredentialCreationResponse.Parse()
	if err != nil {
		return httputil.BadRequest("Invalid credential")
	}

	// Passkeys aren't retrieved from within this function, so we don't need to
	// populate that field on WebAuthnUser.
	cred, err := s.webauthn.CreateCredential(export.WebAuthnUser{User: user}, *hackSession, cc)
	if err != nil {
		return err
	}
	name := body.Name
	if name == "" {
		name = time.Now().Format(time.DateOnly)
	}
	credRaw, _ := json.Marshal(cred)
	if err := s.db.AddPasskey(r.Context(), database.AddPasskeyParams{
		Kthid: user.KTHID,
		Name:  name,
		Data:  string(credRaw),
	}); err != nil {
		return err
	}
	return ""
}

func (s *service) removePasskey(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
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
	if err := s.db.RemovePasskey(r.Context(), database.RemovePasskeyParams{
		Kthid: user.KTHID,
		ID:    passkeyID,
	}); err != nil {
		return err
	}
	return http.RedirectHandler("/account", http.StatusSeeOther)
}
