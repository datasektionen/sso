package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func kthLogin(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if s.RelyingParty == nil {
		return errors.New("KTH OIDC is not reachable at the moment")
	}

	return rp.AuthURLHandler(uuid.NewString, s.RelyingParty)
}

func kthCallback(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if s.RelyingParty == nil {
		return errors.New("KTH OIDC is not reachable at the moment")
	}

	var errMsg string
	if e := r.FormValue("errors"); e != "" {
		errMsg += "\n\n" + e
	}
	if e := r.FormValue("error"); e != "" {
		errMsg += "\n\n" + e
	}
	if errMsg != "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Uh oh! Got an error from KTH:" + errMsg + "\nMaybe try again?"))
		return nil
	}

	return rp.CodeExchangeHandler(func(
		w http.ResponseWriter,
		r *http.Request,
		tokens *oidc.Tokens[*oidc.IDTokenClaims],
		state string,
		rp rp.RelyingParty,
	) {
		kthidAny := tokens.IDTokenClaims.Claims["username"]
		kthid, ok := kthidAny.(string)
		if !ok {
			slog.Error("Could not find a kth-id, or it wasn't a string", "claims", tokens.IDTokenClaims)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		user, err := s.GetUser(r.Context(), kthid)
		if err != nil {
			slog.Error("Could not get user", "kthid", kthid, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if user == nil {
			ok, resp := s.FinishInvite(w, r, kthid)
			if ok {
				httputil.Respond(resp, w, r)
			} else {
				httputil.Respond(templates.MissingAccount(), w, r)
			}
			return
		}
		httputil.Respond(s.LoginUser(r.Context(), user.KTHID), w, r)
	}, s.RelyingParty)
}
