package service

import (
	"context"
	"crypto/rand"
	"log/slog"
	"time"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	oidcHttp "github.com/zitadel/oidc/v3/pkg/http"
)

func initRP(ctx context.Context) (rp.RelyingParty, error) {
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
	return rp, nil
}
