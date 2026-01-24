package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/config"
	"github.com/datasektionen/sso/pkg/email"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
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
	if s.GetLoggedInUser(r) != nil || s.GetLoggedInGuestUser(r) != nil {
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
			Secure:   !config.Config.Dev,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	return templates.Index(s.DevLoginFormOrNilComp)
}

func logout(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return s.Logout(w, r)
}

func account(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if s.GetLoggedInGuestUser(r) != nil {
		return templates.MissingAccount()
	}

	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Redirect("/")
	}
	passkeys, err := s.ListPasskeysForUser(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	pendingEmail, err := s.DB.GetPendingEmailChange(r.Context(), user.KTHID)
	if err != nil && err != pgx.ErrNoRows {
		return nil
	}
	return templates.Account(*user, passkeys, pendingEmail)
}

var yearTagRegex regexp.Regexp = *regexp.MustCompile(`^[A-Z][A-Za-z]{0,3}-\d{2}$`)

func updateAccount(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Invalid form body")
	}
	pendingEmail, err := s.DB.GetPendingEmailChange(r.Context(), user.KTHID)
	if err != nil && err != pgx.ErrNoRows {
		return nil
	}
	yearTagList := r.Form["year-tag"]
	if len(yearTagList) > 0 {
		yearTag := yearTagList[0]
		if !yearTagRegex.Match([]byte(yearTag)) {
			return templates.AccountSettingsForm(*user, pendingEmail, map[string]string{"year-tag": `Invalid format. Must match ` + yearTagRegex.String()})
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
		*user, err = s.UserSetNameChangeRequest(r.Context(), *user, firstNameChangeRequest, familyNameChangeRequest)
		if err != nil {
			return err
		}
	}
	return templates.AccountSettingsForm(*user, pendingEmail, nil)
}

var emailRegex = regexp.MustCompile(`.+@.+\..+`)

func beginEmailChange(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}

	newEmail := strings.TrimSpace(r.FormValue("new-email"))
	if !emailRegex.MatchString(newEmail) {
		return templates.AccountSettingsForm(*user, "", map[string]string{"email": "Invalid email format"})
	}

	code := randomCode()

	if err := s.DB.BeginEmailChange(r.Context(), database.BeginEmailChangeParams{
		Kthid:    user.KTHID,
		NewEmail: newEmail,
		Code:     code,
	}); err != nil {
		return err
	}

	if err := email.Send(
		r.Context(),
		newEmail,
		"SSO - Email Change Verification",
		fmt.Sprintf("Your verification code to change your email address on [Datasektionen SSO](%s) is `%s`", config.Config.Origin, code),
	); err != nil {
		return err
	}

	return templates.AccountSettingsForm(*user, newEmail, nil)
}

func finishEmailChange(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}

	code := r.FormValue("code")
	code, err := parseCode(code)
	if err != nil {
		return err
	}

	res, err := s.DB.FinishEmailChange(r.Context(), database.FinishEmailChangeParams{
		Kthid: user.KTHID,
		Code:  code,
	})
	if err != nil {
		return err
	}

	if res.Ok {
		newUser, err := s.DB.UserSetEmail(r.Context(), database.UserSetEmailParams{
			Kthid: user.KTHID,
			Email: res.NewEmail,
		})
		if err != nil {
			return err
		}
		updatedUser := service.DBUserToModel(newUser)
		return templates.AccountSettingsForm(updatedUser, "", nil)
	}

	pendingEmail, err := s.DB.GetPendingEmailChange(r.Context(), user.KTHID)
	if err != nil && err != pgx.ErrNoRows {
		return nil
	}

	msg := "Code is invalid for an unknown reason. (please tell d-sys)"
	switch res.Reason {
	case "expired":
		msg = "Verification code has expired. Please request a new one."
	case "exhausted":
		msg = "Too many invalid attempts. Please request a new code."
	case "wrong":
		msg = "Invalid code. Please copy-paste or spell better."
	case "no code":
		msg = "No pending email change. Please request one first."
	}
	return templates.AccountSettingsForm(*user, pendingEmail, map[string]string{"email": msg})
}

func cancelEmailChange(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}

	if err := s.DB.CancelEmailChange(r.Context(), user.KTHID); err != nil {
		return err
	}

	return templates.AccountSettingsForm(*user, "", nil)
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
			Secure:   !config.Config.Dev,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		return httputil.Redirect("/oidc/kth/login")
	} else {
		firstName := strings.TrimSpace(r.FormValue("first-name"))
		familyName := strings.TrimSpace(r.FormValue("family-name"))
		emailAddress := strings.TrimSpace(r.FormValue("email"))
		kthid := strings.TrimSpace(r.FormValue("kthid"))

		if kthid == "" {
			kthid = "d-" + strings.ToLower(firstName) + "." + strings.ToLower(familyName)
		}

		if _, err := s.DB.CreateAccountRequestManual(r.Context(), database.CreateAccountRequestManualParams{
			Kthid:      kthid,
			UgKthid:    "d-ug" + kthid,
			Reference:  reference,
			Reason:     reason,
			YearTag:    yearTag,
			FirstName:  firstName,
			FamilyName: familyName,
			Email:      emailAddress,
		}); err != nil {
			return err
		}

		if err := email.Send(
			r.Context(),
			"d-sys@datasektionen.se",
			"Datasektionen Account Requested by "+kthid,
			strings.TrimSpace(fmt.Sprintf(`
				<p>A new account request has been made by %s %s (%s).</p><a href="https://sso.datasektionen.se/admin/account-requests">sso.datasektionen.se/admin/account-requests</a>
			`, firstName, familyName, kthid)),
		); err != nil {
			slog.Error("Could not send account request email", "kthid", kthid, "error", err)
		}

		return httputil.Redirect("/request-account/done")
	}
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
		Secure:   !config.Config.Dev,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	return templates.AcceptInvite()
}
