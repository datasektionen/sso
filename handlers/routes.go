package handlers

import (
	"net/http"

	"github.com/datasektionen/sso/pkg/config"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
)

func MountRoutes(s *service.Service, mux *http.ServeMux, includeInternal bool) {
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Pong! Regards, SSO"))
	})

	// user.go
	mux.Handle("GET /{$}", httputil.Route(s, index))
	mux.Handle("GET /logout", httputil.Route(s, logout))
	mux.Handle("GET /account", httputil.Route(s, account))
	mux.Handle("PATCH /account", httputil.Route(s, updateAccount))
	mux.Handle("GET /request-account", httputil.Route(s, requestAccountPage))
	mux.Handle("POST /request-account", httputil.Route(s, requestAccount))
	mux.Handle("GET /request-account/done", httputil.Route(s, requestAccountDone))
	mux.Handle("GET /invite/{id}", httputil.Route(s, acceptInvite))

	// admin.go
	mux.Handle("GET /admin", authAdmin(s, httputil.Route(s, admin)))

	mux.Handle("GET /admin/members", authAdmin(s, httputil.Route(s, membersPage)))
	mux.Handle("GET /admin/users", authAdmin(s, httputil.Route(s, adminUsersForm)))
	mux.Handle("POST /admin/members/upload-sheet", authAdmin(s, httputil.Route(s, uploadSheet)))
	mux.Handle("GET /admin/members/upload-sheet", authAdmin(s, httputil.Route(s, processSheet)))

	mux.Handle("GET /admin/oidc-clients", authAdmin(s, httputil.Route(s, oidcClients)))
	mux.Handle("POST /admin/oidc-clients", authAdmin(s, httputil.Route(s, createOIDCClient)))
	mux.Handle("PATCH /admin/oidc-clients/{id}", authAdmin(s, httputil.Route(s, updateOIDCClient)))
	mux.Handle("DELETE /admin/oidc-clients/{id}", authAdmin(s, httputil.Route(s, deleteOIDCClient)))
	mux.Handle("POST /admin/oidc-clients/{id}/redirect-uris", authAdmin(s, httputil.Route(s, addRedirectURI)))
	mux.Handle("DELETE /admin/oidc-clients/{id}/redirect-uris/{uri}", authAdmin(s, httputil.Route(s, removeRedirectURI)))

	mux.Handle("GET /admin/invites", authAdmin(s, httputil.Route(s, invites)))
	mux.Handle("GET /admin/invites/{id}", authAdmin(s, httputil.Route(s, invite)))
	mux.Handle("POST /admin/invites", authAdmin(s, httputil.Route(s, createInvite)))
	mux.Handle("DELETE /admin/invites/{id}", authAdmin(s, httputil.Route(s, deleteInvite)))
	mux.Handle("GET /admin/invites/{id}/edit", authAdmin(s, httputil.Route(s, editInviteForm)))
	mux.Handle("PUT /admin/invites/{id}", authAdmin(s, httputil.Route(s, updateInvite)))

	mux.Handle("GET /admin/account-requests", authAdmin(s, httputil.Route(s, accountRequests)))
	mux.Handle("DELETE /admin/account-requests/{id}", authAdmin(s, httputil.Route(s, denyAccountRequest)))
	mux.Handle("POST /admin/account-requests/{id}", authAdmin(s, httputil.Route(s, approveAccountRequest)))

	// dev.go
	if config.Config.Dev {
		mux.Handle("POST /login/dev", httputil.Route(s, devLogin))
		mux.Handle("GET /dev/auto-reload", httputil.Route(s, autoReload))
	}

	// passkey.go
	mux.Handle("POST /login/passkey/begin", httputil.Route(s, beginLoginPasskey))
	mux.Handle("POST /login/passkey/finish", httputil.Route(s, finishLoginPasskey))

	mux.Handle("GET /passkey/add-form", httputil.Route(s, addPasskeyForm))
	mux.Handle("POST /passkey", httputil.Route(s, addPasskey))
	mux.Handle("DELETE /passkey/{id}", httputil.Route(s, removePasskey))

	// oidcrp.go
	mux.Handle("GET /oidc/kth/login", httputil.Route(s, kthLogin))
	mux.Handle("GET /oidc/kth/callback", httputil.Route(s, kthCallback))

	// legacyapi.go
	mux.Handle("/legacyapi/hello", httputil.Route(s, legacyAPIHello))
	mux.Handle("/legacyapi/login", httputil.Route(s, legacyAPILogin))
	mux.Handle("/legacyapi/callback", httputil.Route(s, legacyAPICallback))
	mux.Handle("/legacyapi/verify/{token}", httputil.Route(s, legacyAPIVerify))
	mux.Handle("/legacyapi/logout", httputil.Route(s, legacyAPILogout))

	// internalapi.go
	if includeInternal {
		mux.Handle("GET /api/users", httputil.Route(s, apiListUsers))
	}
}
