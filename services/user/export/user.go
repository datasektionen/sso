package export

import (
	"context"
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
)

type Service interface {
	GetUser(ctx context.Context, kthid string) (*User, error)
	LoginUser(ctx context.Context, kthid string) httputil.ToResponse
	GetLoggedInUser(r *http.Request) (*User, error)
	Logout(w http.ResponseWriter, r *http.Request) httputil.ToResponse
}

type User struct {
	KTHID      string `db:"kthid"`
	WebAuthnID []byte `db:"webauthn_id"`
}
