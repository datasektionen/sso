package handlers

import (
	"net/http"

	"github.com/datasektionen/logout/pkg/config"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/service"
)

func MountRoutes(s *service.Service) {
	http.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Pong! Regards, Logout"))
	})

	// user.go
	http.Handle("GET /{$}", httputil.Route(s, index))
	http.Handle("GET /logout", httputil.Route(s, logout))
	http.Handle("GET /account", httputil.Route(s, account))
	http.Handle("PATCH /account", httputil.Route(s, updateAccount))
	http.Handle("GET /invite/{id}", httputil.Route(s, acceptInvite))

	// admin.go
	http.Handle("GET /admin", authAdmin(s, httputil.Route(s, admin)))

	http.Handle("GET /admin/members", authAdmin(s, httputil.Route(s, membersPage)))
	http.Handle("GET /admin/users", authAdmin(s, httputil.Route(s, adminUsersForm)))
	http.Handle("POST /admin/members/upload-sheet", authAdmin(s, httputil.Route(s, uploadSheet)))
	http.Handle("GET /admin/members/upload-sheet", authAdmin(s, httputil.Route(s, processSheet)))

	http.Handle("GET /admin/oidc-clients", authAdmin(s, httputil.Route(s, oidcClients)))
	http.Handle("POST /admin/oidc-clients", authAdmin(s, httputil.Route(s, createOIDCClient)))
	http.Handle("DELETE /admin/oidc-clients/{id}", authAdmin(s, httputil.Route(s, deleteOIDCClient)))
	http.Handle("POST /admin/oidc-clients/{id}", authAdmin(s, httputil.Route(s, addRedirectURI)))
	http.Handle("DELETE /admin/oidc-clients/{id}/{uri}", authAdmin(s, httputil.Route(s, removeRedirectURI)))

	http.Handle("GET /admin/invites", authAdmin(s, httputil.Route(s, invites)))
	http.Handle("GET /admin/invites/{id}", authAdmin(s, httputil.Route(s, invite)))
	http.Handle("POST /admin/invites", authAdmin(s, httputil.Route(s, createInvite)))
	http.Handle("DELETE /admin/invites/{id}", authAdmin(s, httputil.Route(s, deleteInvite)))
	http.Handle("GET /admin/invites/{id}/edit", authAdmin(s, httputil.Route(s, editInviteForm)))
	http.Handle("PUT /admin/invites/{id}", authAdmin(s, httputil.Route(s, updateInvite)))

	// dev.go
	if config.Config.Dev {
		http.Handle("POST /login/dev", httputil.Route(s, devLogin))
		http.Handle("GET /dev/auto-reload", httputil.Route(s, autoReload))
	}

	// passkey.go
	http.Handle("POST /login/passkey/begin", httputil.Route(s, beginLoginPasskey))
	http.Handle("POST /login/passkey/finish", httputil.Route(s, finishLoginPasskey))

	http.Handle("GET /passkey/add-form", httputil.Route(s, addPasskeyForm))
	http.Handle("POST /passkey", httputil.Route(s, addPasskey))
	http.Handle("DELETE /passkey/{id}", httputil.Route(s, removePasskey))

	// oidcrp.go
	http.Handle("GET /login/oidc/kth", httputil.Route(s, kthLogin))
	http.Handle("GET /oidc/kth/callback", httputil.Route(s, kthCallback))

	// legacyapi.go
	http.Handle("/legacyapi/hello", httputil.Route(s, legacyAPIHello))
	http.Handle("/legacyapi/login", httputil.Route(s, legacyAPILogin))
	http.Handle("/legacyapi/callback", httputil.Route(s, legacyAPICallback))
	http.Handle("/legacyapi/verify/{token}", httputil.Route(s, legacyAPIVerify))
	http.Handle("/legacyapi/logout", httputil.Route(s, legacyAPILogout))
}
