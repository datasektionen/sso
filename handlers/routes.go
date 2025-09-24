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
	mux.Handle("GET /admin/members", authorize(s, httputil.Route(s, membersPage), "read-members", nil))
	mux.Handle("GET /admin/users", authorize(s, httputil.Route(s, adminUsersForm), "read-members", nil))
	mux.Handle("POST /admin/members/upload-sheet", authorize(s, httputil.Route(s, uploadSheet), "write-members", nil))
	mux.Handle("GET /admin/members/upload-sheet", authorize(s, httputil.Route(s, processSheet), "write-members", nil))

	mux.Handle("GET /admin/oidc-clients", authorize(s, httputil.Route(s, oidcClients), "read-oidc-clients", nil))
	mux.Handle("POST /admin/oidc-clients", authorize(s, httputil.Route(s, createOIDCClient), "write-oidc-clients", func(r *http.Request) string { return r.FormValue("id") }))
	mux.Handle("PATCH /admin/oidc-clients/{id}", authorize(s, httputil.Route(s, updateOIDCClient), "write-oidc-clients", func(r *http.Request) string { return r.PathValue("id") }))
	mux.Handle("DELETE /admin/oidc-clients/{id}", authorize(s, httputil.Route(s, deleteOIDCClient), "write-oidc-clients", func(r *http.Request) string { return r.PathValue("id") }))
	mux.Handle("POST /admin/oidc-clients/{id}/redirect-uris", authorize(s, httputil.Route(s, addRedirectURI), "write-oidc-clients", func(r *http.Request) string { return r.PathValue("id") }))
	mux.Handle("DELETE /admin/oidc-clients/{id}/redirect-uris/{uri}", authorize(s, httputil.Route(s, removeRedirectURI), "write-oidc-clients", func(r *http.Request) string { return r.PathValue("id") }))

	mux.Handle("GET /admin/invites", authorize(s, httputil.Route(s, invites), "read-invites", nil))
	mux.Handle("GET /admin/invites/{id}", authorize(s, httputil.Route(s, invite), "read-invites", nil))
	mux.Handle("POST /admin/invites", authorize(s, httputil.Route(s, createInvite), "write-invites", nil))
	mux.Handle("DELETE /admin/invites/{id}", authorize(s, httputil.Route(s, deleteInvite), "write-invites", nil))
	mux.Handle("GET /admin/invites/{id}/edit", authorize(s, httputil.Route(s, editInviteForm), "write-invites", nil))
	mux.Handle("PUT /admin/invites/{id}", authorize(s, httputil.Route(s, updateInvite), "write-invites", nil))

	mux.Handle("GET /admin/account-requests", authorize(s, httputil.Route(s, accountRequests), "read-account-requests", nil))
	mux.Handle("DELETE /admin/account-requests/{id}", authorize(s, httputil.Route(s, denyAccountRequest), "manage-account-requests", nil))
	mux.Handle("POST /admin/account-requests/{id}", authorize(s, httputil.Route(s, approveAccountRequest), "manage-account-requests", nil))

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
		mux.Handle("GET /api/search", httputil.Route(s, apiSearchUsers))
	}
}
