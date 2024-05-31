package oidcrp

import (
	"log/slog"
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func (s *service) kthLogin(r *http.Request) httputil.ToResponse {
	return rp.AuthURLHandler(uuid.NewString, s.relyingParty)
}

func (s *service) kthCallback(r *http.Request) httputil.ToResponse {
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
		user, err := s.user.GetUser(r.Context(), kthid)
		if err != nil {
			slog.Error("Could not get user", "kthid", user.KTHID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if user == nil {
			// TODO: show user creation request note/thingie
			httputil.Forbidden().(httputil.HttpError).ServeHTTP(w, r)
			return
		}
		sessionID, err := s.user.CreateSession(r.Context(), user.KTHID)
		if err != nil {
			slog.Error("Could not create session", "kthid", user.KTHID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: sessionID.String(),
			Path:  "/",
		})
		http.Redirect(w, r, "/account", http.StatusSeeOther)
	}, s.relyingParty)
}
