package handlers

import (
	"net/http"
	"regexp"
	"time"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/pkg/pls"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const nextUrlCookie string = "_sso_next-url"

func validNextURL(url string) bool {
	if url == "" {
		return true
	}
	// The "url" must be a path (and possibly params)
	if len(url) > 0 && url[0] != '/' {
		return false
	}
	// If it starts with two slashes there will be an implicit `https:` in front, so then it's not a path
	if len(url) > 1 && url[1] == '/' {
		return false
	}
	return true
}

func index(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	nextURL := r.FormValue("next-url")
	if !validNextURL(nextURL) {
		return httputil.BadRequest("Invalid return url")
	}
	hasCookie := false
	if nextURL == "" {
		c, _ := r.Cookie(nextUrlCookie)
		if c != nil {
			nextURL = c.Value
			hasCookie = true
		}
	}
	if nextURL == "" {
		nextURL = "/account"
	}
	if kthid, err := s.GetLoggedInKTHID(r); err != nil {
		return err
	} else if kthid != "" {
		if hasCookie {
			http.SetCookie(w, &http.Cookie{Name: nextUrlCookie, MaxAge: -1})
		}
		http.Redirect(w, r, nextURL, http.StatusSeeOther)
		return nil
	}
	if nextURL != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     nextUrlCookie,
			Value:    nextURL,
			MaxAge:   int((time.Minute * 10).Seconds()),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	return templates.Index(s.DevLoginForm)
}

func logout(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return s.Logout(w, r)
}

func account(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	passkeys, err := s.ListPasskeysForUser(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	isAdmin, err := pls.CheckUser(r.Context(), user.KTHID, "admin-read")
	if err != nil {
		return err
	}
	return templates.Account(*user, passkeys, isAdmin)
}

var yearTagRegex regexp.Regexp = *regexp.MustCompile(`^[A-Z][A-Za-z]{0,3}-\d{2}$`)

func updateAccount(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user, err := s.GetLoggedInUser(r)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.Unauthorized()
	}
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Invalid form body")
	}
	yearTagList := r.Form["year-tag"]
	if len(yearTagList) > 0 {
		yearTag := yearTagList[0]
		if !yearTagRegex.Match([]byte(yearTag)) {
			return templates.AccountSettingsForm(*user, map[string]string{"year-tag": `Invalid format. Must match ` + yearTagRegex.String()})
		}
		var err error
		*user, err = s.UserSetYear(r.Context(), user.KTHID, yearTag)
		if err != nil {
			return err
		}
	}
	var firstNameChangeRequest, familyNameChangeRequest string
	var doNameChangeRequest bool
	if n := r.Form["first-name"]; len(n) != 0 {
		firstNameChangeRequest = n[0]
		doNameChangeRequest = true
	}
	if n := r.Form["family-name"]; len(n) != 0 {
		familyNameChangeRequest = n[0]
		doNameChangeRequest = true
	}
	if doNameChangeRequest {
		var err error
		*user, err = s.UserSetNameChangeRequest(r.Context(), user.KTHID, firstNameChangeRequest, familyNameChangeRequest)
		if err != nil {
			return err
		}
	}
	return templates.AccountSettingsForm(*user, nil)
}

func requestAccountPage(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return templates.RequestAccount()
}

func requestAccount(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	var hasKTHAccount bool
	switch r.FormValue("have-kth-account") {
	case "yes":
		hasKTHAccount = true
	case "no":
		hasKTHAccount = false
	default:
		return httputil.BadRequest("Invalid `have-kth-account`. Must be `yes` or `no`")
	}

	reference := r.FormValue("reference")
	reason := r.FormValue("reason")
	yearTag := r.FormValue("year-tag")

	if hasKTHAccount {
		id, err := s.DB.CreateAccountRequest(r.Context(), database.CreateAccountRequestParams{
			Reference: reference,
			Reason:    reason,
			YearTag:   yearTag,
		})
		if err != nil {
			return err
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "account-request-id",
			Value:    id.String(),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		return httputil.Redirect("/oidc/kth/login")
	} else {
		panic("todo")
	}

	// TODO: send email to admins
}

func requestAccountDone(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return templates.AccountRequestDone()
}

func acceptInvite(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	idString := r.PathValue("id")
	if idString == "-" {
		idCookie, _ := r.Cookie("invite")
		if idCookie == nil {
			return httputil.BadRequest("No invite id found")
		}
		idString = idCookie.Value
	}
	id, err := uuid.Parse(idString)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	inv, err := s.DB.GetInvite(r.Context(), id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such invite")
	} else if err != nil {
		return err
	}
	if time.Now().After(inv.ExpiresAt.Time) {
		return httputil.BadRequest("Invite expired")
	}
	if inv.MaxUses.Valid && inv.CurrentUses >= inv.MaxUses.Int32 {
		return httputil.BadRequest("This invite cannot be used to create more users")
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "invite",
		Value:    id.String(),
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	return templates.AcceptInvite()
}
