package oidcrp

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/service"
	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func MountRoutes(s *service.Service) {
	http.Handle("GET /login/oidc/kth", httputil.Route(s, kthLogin))
	http.Handle("GET /oidc/kth/callback", httputil.Route(s, kthCallback))
}

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
				// TODO: show a better user creation request note/thingie
				httputil.Respond(httputil.Forbidden(
					"Your KTH account is not connected to a Datasektionen account. This should happen "+
						"automatically if you are a chapter member and otherwise you must receive an invitation. If "+
						"you believe this is a mistake, please contact head of IT at d-sys@datasektionen.se",
				), w, r)
			}
			return
		}
		httputil.Respond(s.LoginUser(r.Context(), user.KTHID), w, r)
	}, s.RelyingParty)
}
