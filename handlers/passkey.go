package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/models"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

func beginLoginPasskey(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	kthid := r.FormValue("kthid")
	user, err := s.GetUser(r.Context(), kthid)
	if err != nil {
		return err
	}
	if user == nil {
		w.Header().Add("HX-Reswap", "beforeend")
		return `<p class="error">No such user</p>`
	}
	passkeys, err := s.ListPasskeysForUser(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	if len(passkeys) == 0 {
		w.Header().Add("HX-Reswap", "beforeend")
		return `<p class="error">You have no registered passkeys</p>`
	}
	credAss, sessionData, err := s.WebAuthn.BeginLogin(models.WebAuthnUser{User: user, Passkeys: passkeys})
	if err != nil {
		return err
	}
	sessionDataBytes, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}
	sessionID, err := s.DB.StoreWebAuthnSessionData(r.Context(), database.StoreWebAuthnSessionDataParams{Data: sessionDataBytes, Kthid: kthid})
	if err != nil {
		return err
	}
	return templates.PasskeyLoginForm(kthid, credAss, sessionID)
}

func finishLoginPasskey(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	var body struct {
		SessionID uuid.UUID                            `json:"session"`
		Cred      protocol.CredentialAssertionResponse `json:"cred"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return httputil.BadRequest("Invalid credential")
	}
	credAss, err := body.Cred.Parse()
	if err != nil {
		return httputil.BadRequest("Invalid credential")
	}

	session, err := s.DB.TakeWebAuthnSessionData(r.Context(), body.SessionID)
	if err != nil {
		return err
	}
	var sessionData webauthn.SessionData
	if err := json.Unmarshal(session.Data, &sessionData); err != nil {
		return err
	}

	user, err := s.GetUser(r.Context(), session.Kthid)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}
	passkeys, err := s.ListPasskeysForUser(r.Context(), user.KTHID)
	if err != nil {
		return err
	}

	_, err = s.WebAuthn.ValidateLogin(models.WebAuthnUser{User: user, Passkeys: passkeys}, sessionData, credAss)
	if err != nil {
		return err
	}

	return s.LoginUser(r.Context(), user.KTHID)
}

func addPasskeyForm(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}
	creation, sessionData, err := s.WebAuthn.BeginRegistration(models.WebAuthnUser{User: user})
	if err != nil {
		return err
	}
	sessionDataBytes, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}
	sessionID, err := s.DB.StoreWebAuthnSessionData(r.Context(), database.StoreWebAuthnSessionDataParams{Data: sessionDataBytes, Kthid: user.KTHID})
	if err != nil {
		return err
	}

	return templates.AddPasskeyForm(creation, sessionID)
}

func addPasskey(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}

	var body struct {
		Name      string    `json:"name"`
		SessionID uuid.UUID `json:"session"`
		protocol.CredentialCreationResponse
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return httputil.BadRequest("Invalid credential")
	}
	cc, err := body.CredentialCreationResponse.Parse()
	if err != nil {
		return httputil.BadRequest("Invalid credential")
	}

	session, err := s.DB.TakeWebAuthnSessionData(r.Context(), body.SessionID)
	if err != nil {
		return err
	}
	var sessionData webauthn.SessionData
	if err := json.Unmarshal(session.Data, &sessionData); err != nil {
		return err
	}
	// Passkeys aren't retrieved from within this function, so we don't need to
	// populate that field on WebAuthnUser.
	cred, err := s.WebAuthn.CreateCredential(models.WebAuthnUser{User: user}, sessionData, cc)
	if err != nil {
		return err
	}
	name := body.Name
	if name == "" {
		name = time.Now().Format(time.DateOnly)
	}
	credRaw, _ := json.Marshal(cred)
	id, err := s.DB.AddPasskey(r.Context(), database.AddPasskeyParams{
		Kthid: user.KTHID,
		Name:  name,
		Data:  string(credRaw),
	})
	if err != nil {
		return err
	}
	passkey := models.Passkey{ID: id, Name: name}
	return templates.ShowPasskey(passkey)
}

func removePasskey(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}
	passkeyID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	if err := s.DB.RemovePasskey(r.Context(), database.RemovePasskeyParams{
		Kthid: user.KTHID,
		ID:    passkeyID,
	}); err != nil {
		return err
	}
	return nil
}
