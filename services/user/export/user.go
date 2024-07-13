package export

import (
	"context"
	"net/http"
	"time"

	"github.com/datasektionen/logout/pkg/httputil"
)

type Service interface {
	GetUser(ctx context.Context, kthid string) (*User, error)
	LoginUser(ctx context.Context, kthid string) httputil.ToResponse
	GetLoggedInKTHID(r *http.Request) (string, error)
	GetLoggedInUser(r *http.Request) (*User, error)
	Logout(w http.ResponseWriter, r *http.Request) httputil.ToResponse
	FinishInvite(w http.ResponseWriter, r *http.Request, kthid string) (bool, httputil.ToResponse)
}

type User struct {
	KTHID      string
	UGKTHID    string
	Email      string
	FirstName  string
	FamilyName string
	YearTag    string
	MemberTo   time.Time
	WebAuthnID []byte
}
