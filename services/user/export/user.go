package export

import (
	"context"
	"net/http"
	"net/url"

	"github.com/datasektionen/logout/models"
	"github.com/datasektionen/logout/pkg/httputil"
)

type Service interface {
	GetUser(ctx context.Context, kthid string) (*models.User, error)
	LoginUser(ctx context.Context, kthid string) httputil.ToResponse
	GetLoggedInKTHID(r *http.Request) (string, error)
	GetLoggedInUser(r *http.Request) (*models.User, error)
	Logout(w http.ResponseWriter, r *http.Request) httputil.ToResponse
	FinishInvite(w http.ResponseWriter, r *http.Request, kthid string) (bool, httputil.ToResponse)
	RedirectToLogin(w http.ResponseWriter, r *http.Request, nextURL string)
}

func LoginPath(nextURL string) string {
	return "/?" + url.Values{"next-url": []string{nextURL}}.Encode()
}
