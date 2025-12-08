package handlers

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/email"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 6

func beginLoginEmail(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	ctx := context.Background()
	kthid := r.FormValue("kthid")
	user, err := s.DB.GetUser(ctx, kthid)

	if err != nil {
		return err
	}

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, codeLength)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	code := string(b)

	err = s.DB.BeginEmailLogin(ctx, database.BeginEmailLoginParams{Kthid: kthid, Code: code})

	if err != nil {
		return err
	}

	err = email.Send(
		ctx,
		user.Email,
		"SSO - Login Code",
		strings.TrimSpace(`Your temporary login code is *`+code+`*`))

	if err != nil {
		return err
	}

	return templates.CodeForm(kthid)
}

func finishLoginEmail(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	ctx := context.Background()
	kthid := r.FormValue("kthid")
	code := r.FormValue("code")

	code_info, err := s.DB.GetEmailLogin(ctx, database.GetEmailLoginParams{Kthid: kthid, Code: code})
	if err != nil {
		return err
	}

	if code_info.Attempts > 3 {
		s.DB.ClearEmailCodes(ctx, kthid)
		return httputil.Unauthorized()
	}

	if code_info.CreationTime.Add(time.Duration(1000 * 1000 * 60 * 10)).Compare(time.Now()) <= 0 { // 10 minutes
		s.DB.ClearEmailCodes(ctx, kthid)
		return s.LoginUser(r.Context(), kthid, true)
	} else {
		s.DB.IncreaseAttempts(ctx, kthid)
		return httputil.Unauthorized()
	}
}
