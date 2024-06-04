package oidcrp

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"time"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/httputil"
	user "github.com/datasektionen/logout/services/user/export"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	oidcHttp "github.com/zitadel/oidc/v3/pkg/http"
)

type service struct {
	relyingParty rp.RelyingParty
	user         user.Service
}

func NewService(ctx context.Context) (*service, error) {
	// TODO: persist?
	hashKey := make([]byte, 32)
	if _, err := rand.Read(hashKey); err != nil {
		return nil, err
	}
	encryptKey := make([]byte, 32)
	if _, err := rand.Read(encryptKey); err != nil {
		return nil, err
	}
	rp, err := rp.NewRelyingPartyOIDC(
		ctx,
		config.Config.KTHOIDCIssuerURL.String(),
		config.Config.KTHOIDCClientID,
		config.Config.KTHOIDCClientSecret,
		config.Config.KTHOIDCRPOrigin.String()+"/oidc/kth/callback",
		[]string{"openid", "profile", "email"},
		rp.WithCookieHandler(oidcHttp.NewCookieHandler(
			hashKey,
			encryptKey,
			oidcHttp.WithMaxAge(int((10*time.Minute).Seconds())),
		)),
	)
	if err != nil {
		slog.Error("Error initializing relying party for KTH's OIDC provider", "error", err)
		rp = nil // Should already be the case but doesn't hurt to make sure
	}

	s := &service{relyingParty: rp}

	http.Handle("GET /login/oidc/kth", httputil.Route(s.kthLogin))
	http.Handle("GET /oidc/kth/callback", httputil.Route(s.kthCallback))

	return s, nil
}

func (s *service) Assign(user user.Service) {
	s.user = user
}
