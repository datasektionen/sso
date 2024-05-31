package export

import (
	"context"
	"net/http"

	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/google/uuid"
)

type Service interface {
	GetUser(ctx context.Context, kthid string) (*User, error)
	CreateSession(ctx context.Context, kthid string) (uuid.UUID, error)
	GetLoggedInUser(r *http.Request) (*User, error)
	Logout(r *http.Request) httputil.ToResponse
}

type User struct {
	KTHID      string `db:"kthid"`
	WebAuthnID []byte `db:"webauthn_id"`
}
