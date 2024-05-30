package main

import (
	"log/slog"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type Service struct {
	webauthn *webauthn.WebAuthn
	db       DB
}

func NewService() (*Service, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          "localhost",
		RPDisplayName: "Konglig Datasektionen",
		RPOrigins:     []string{"http://localhost:3000"},
	})
	if err != nil {
		return nil, err
	}
	db, err := NewDB("postgresql://logout:logout@localhost:5432/logout")
	if err != nil {
		return nil, err
	}
	return &Service{webauthn: wa, db: db}, nil
}

func (s *Service) Index(r *http.Request) ToResponse {
	user, err := s.getLoggedInUser(r)
	if err != nil {
		return err
	}
	return index(user)
}

func (s *Service) Logout(r *http.Request) ToResponse {
	sessionCookie, _ := r.Cookie("session")
	if sessionCookie != nil {
		if err := s.db.RemoveSession(r.Context(), sessionCookie.Value); err != nil {
			return err
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func (s *Service) Account(r *http.Request) ToResponse {
	user, err := s.getLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return Unauthorized()
	}
	return account(*user)
}

func (s *Service) ShowRegister(r *http.Request) ToResponse {
	return register()
}

func (s *Service) DoRegister(r *http.Request) ToResponse {
	kthid := r.FormValue("kthid")
	if len(kthid) < 2 {
		return BadRequest("Invalid kthid")
	}
	if err := s.db.CreateUser(r.Context(), kthid); err != nil {
		return err
	}
	slog.Info("User registrated", "kthid", kthid)
	return http.RedirectHandler("/", http.StatusSeeOther)
}

func (s *Service) ShowLoginDev(r *http.Request) ToResponse {
	return loginDev()
}

func (s *Service) DoLoginDev(r *http.Request) ToResponse {
	user, err := s.db.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return BadRequest("No such user")
	}
	sessionID, err := s.db.CreateSession(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: sessionID.String(),
			Path:  "/",
		})
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	})
}

var hackSession *webauthn.SessionData

func (s *Service) BeginLoginPasskey(r *http.Request) ToResponse {
	user, err := s.db.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return BadRequest("No such user")
	}
	if len(user.Passkeys) == 0 {
		return BadRequest("You have no registered passkeys")
	}
	credAss, sessionData, err := s.webauthn.BeginLogin(WebAuthnUser{user})
	hackSession = sessionData
	if err != nil {
		return err
	}
	return loginPasskey(credAss)
}

func (s *Service) FinishLoginPasskey(r *http.Request) ToResponse {
	user, err := s.db.GetUser(r.Context(), r.FormValue("kthid"))
	if err != nil {
		return err
	}
	if user == nil {
		return BadRequest("No such user")
	}
	_, err = s.webauthn.FinishLogin(WebAuthnUser{user}, *hackSession, r)
	if err != nil {
		return err
	}
	sessionID, err := s.db.CreateSession(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: sessionID.String(),
			Path:  "/",
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func (s *Service) BeginAddPasskey(r *http.Request) ToResponse {
	user, err := s.getLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return Unauthorized()
	}
	creation, sessionData, err := s.webauthn.BeginRegistration(WebAuthnUser{user})
	if err != nil {
		return err
	}
	hackSession = sessionData

	return addPasskey(creation)
}

func (s *Service) FinishAddPasskey(r *http.Request) ToResponse {
	user, err := s.getLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return Unauthorized()
	}
	cred, err := s.webauthn.FinishRegistration(WebAuthnUser{user}, *hackSession, r)
	if err != nil {
		return err
	}
	name := r.FormValue("name")
	if err := s.db.AddPasskey(r.Context(), user.KTHID, name, *cred); err != nil {
		return err
	}
	return ""
}

func (s *Service) RemovePasskey(r *http.Request) ToResponse {
	user, err := s.getLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return Unauthorized()
	}
	passkeyID, err := uuid.Parse(r.FormValue("passkey-id"))
	if err != nil {
		return BadRequest("Invalid uuid")
	}
	if err := s.db.RemovePasskey(r.Context(), user.KTHID, passkeyID); err != nil {
		return err
	}
	return http.RedirectHandler("/account", http.StatusSeeOther)
}


func (s *Service) getLoggedInKTHID(r *http.Request) (string, error) {
	sessionCookie, _ := r.Cookie("session")
	if sessionCookie == nil {
		return "", nil
	}
	return s.db.GetSession(sessionCookie.Value)
}

func (s *Service) getLoggedInUser(r *http.Request) (*User, error) {
	kthid, err := s.getLoggedInKTHID(r)
	if err != nil {
		return nil, err
	}
	user, err := s.db.GetUser(r.Context(), kthid)
	return user, nil
}
