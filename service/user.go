package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/models"
	"github.com/datasektionen/sso/pkg/auth"
	"github.com/datasektionen/sso/pkg/email"
	"github.com/datasektionen/sso/pkg/hive"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/pkg/kthldap"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func dbUserToModel(user database.User) models.User {
	var memberTo time.Time
	if user.MemberTo.Valid {
		memberTo = user.MemberTo.Time
	}
	return models.User{
		KTHID:                   user.Kthid,
		UGKTHID:                 user.UgKthid,
		Email:                   user.Email,
		FirstName:               user.FirstName,
		FamilyName:              user.FamilyName,
		YearTag:                 user.YearTag,
		MemberTo:                memberTo,
		WebAuthnID:              user.WebauthnID,
		FirstNameChangeRequest:  user.FirstNameChangeRequest,
		FamilyNameChangeRequest: user.FamilyNameChangeRequest,
	}
}

func DBUsersToModel(users []database.User) []models.User {
	us := make([]models.User, len(users))
	for i, u := range users {
		us[i] = dbUserToModel(u)
	}
	return us
}

func (s *Service) GetUser(ctx context.Context, kthid string) (*models.User, error) {
	user, err := s.DB.GetUser(ctx, kthid)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u := dbUserToModel(user)
	return &u, nil
}

func (s *Service) UserSetYear(ctx context.Context, kthid string, yearTag string) (models.User, error) {
	user, err := s.DB.UserSetYear(ctx, database.UserSetYearParams{
		Kthid:   kthid,
		YearTag: yearTag,
	})
	if err != nil {
		return models.User{}, err
	}
	return dbUserToModel(user), nil
}

func (s *Service) UserSetNameChangeRequest(ctx context.Context, user models.User, firstName string, familyName string) (models.User, error) {
	newUser, err := s.DB.UserSetNameChangeRequest(ctx, database.UserSetNameChangeRequestParams{
		Kthid:                   user.KTHID,
		FirstNameChangeRequest:  firstName,
		FamilyNameChangeRequest: familyName,
	})
	if err != nil {
		return models.User{}, err
	}
	if err := email.Send(ctx, "d-sys@datasektionen.se", "SSO - Name change requested", strings.TrimSpace(`
The user `+user.FirstName+` `+user.FamilyName+` (`+user.KTHID+`) has requested to change their name to `+"`"+firstName+"`"+` `+"`"+familyName+"`"+`.
	`)); err != nil {
		return models.User{}, err
	}
	return dbUserToModel(newUser), nil
}

func (s *Service) LoginUser(ctx context.Context, kthid string, redirect bool) httputil.ToResponse {
	perms, err := hive.GetSSOPermissions(ctx, kthid)
	if err != nil {
		return err
	}
	jsonPerms, err := json.Marshal(perms)
	if err != nil {
		slog.Error("LoginUser: could not marshal permissions. Resorting to empty permissions", "permissions", perms, "error", err)
		jsonPerms = []byte(`{}`)
	}
	sessionID, err := s.DB.CreateSession(ctx, database.CreateSessionParams{
		Kthid:       pgtype.Text{String: kthid, Valid: true},
		Permissions: jsonPerms,
	})
	if err != nil {
		return err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, auth.SessionCookie(sessionID.String()))
		if redirect {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

func (s *Service) LoginGuestUser(ctx context.Context, guestUser models.GuestUser, redirect bool) httputil.ToResponse {
	guestData, err := json.Marshal(guestUser)
	if err != nil {
		return err
	}
	sessionID, err := s.DB.CreateGuestSession(ctx, database.CreateGuestSessionParams{
		GuestData:   guestData,
		Permissions: []byte(`{}`),
	})
	if err != nil {
		return err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, auth.SessionCookie(sessionID.String()))
		if redirect {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
}

func (s *Service) WithSession(r *http.Request) (*http.Request, error) {
	sessionCookie, _ := r.Cookie(auth.SessionCookieName)
	if sessionCookie == nil {
		return r, nil
	}
	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		return r, nil
	}
	session, err := s.DB.GetSession(r.Context(), sessionID)
	if err == pgx.ErrNoRows {
		return r, nil
	}
	if err != nil {
		return r, err
	}
	var permissions hive.Permissions
	err = json.Unmarshal(session.Permissions, &permissions)
	if err != nil {
		slog.Error("Got invalid json in permissions in database. Defaulting to empty permission set.", "permissions", session.Permissions, "error", err)
	}

	ctx := context.WithValue(r.Context(), hive.PermissionsCtxKey{}, permissions)

	if session.Kthid.Valid {
		user, err := s.GetUser(r.Context(), session.Kthid.String)
		if err != nil {
			return r, err
		}
		ctx = context.WithValue(ctx, models.UserCtxKey{}, user)
	} else if len(session.GuestData) > 0 {
		var guestUser models.GuestUser
		err = json.Unmarshal(session.GuestData, &guestUser)
		if err != nil {
			slog.Error("Got invalid json in guest_data in database.", "guest_data", session.GuestData, "error", err)
			return r, nil
		}
		ctx = context.WithValue(ctx, models.GuestUserCtxKey{}, &guestUser)
	}

	return r.WithContext(ctx), nil
}

func (s *Service) GetLoggedInUser(r *http.Request) *models.User {
	user := r.Context().Value(models.UserCtxKey{})
	if user == nil {
		return nil
	}
	return user.(*models.User)
}

func (s *Service) GetLoggedInGuestUser(r *http.Request) *models.GuestUser {
	guestUser := r.Context().Value(models.GuestUserCtxKey{})
	if guestUser == nil {
		return nil
	}
	return guestUser.(*models.GuestUser)
}

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	sessionCookie, _ := r.Cookie(auth.SessionCookieName)
	if sessionCookie != nil {
		sessionID, err := uuid.Parse(sessionCookie.Value)
		if err != nil {
			if err := s.DB.RemoveSession(r.Context(), sessionID); err != nil {
				return err
			}
		}
	}
	http.SetCookie(w, &http.Cookie{Name: auth.SessionCookieName, MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func (s *Service) RedirectToLogin(w http.ResponseWriter, r *http.Request, nextURL string) {
	http.Redirect(w, r, "/?"+url.Values{"next-url": {nextURL}}.Encode(), http.StatusSeeOther)
}

func (s *Service) FinishAccountRequestKTH(w http.ResponseWriter, r *http.Request, kthid string) httputil.ToResponse {
	requestIDCookie, _ := r.Cookie("account-request-id")
	if requestIDCookie == nil {
		return nil
	}
	requestID, err := uuid.Parse(requestIDCookie.Value)
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	if err := s.DB.FinishAccountRequestKTH(r.Context(), database.FinishAccountRequestKTHParams{
		ID:    requestID,
		Kthid: pgtype.Text{String: kthid, Valid: true},
	}); err != nil {
		return err
	}

	person, err := kthldap.Lookup(r.Context(), kthid)
	if err != nil {
		return err
	}
	if err := email.Send(
		r.Context(),
		"d-sys@datasektionen.se",
		"Datasektionen Account Requested by "+person.KTHID,
		strings.TrimSpace(fmt.Sprintf(`
			<p>A new account request has been made by %s %s (%s).</p><a href="https://sso.datasektionen.se/admin/account-requests">sso.datasektionen.se/admin/account-requests</a>
		`, person.FirstName, person.FamilyName, person.KTHID)),
	); err != nil {
		return err
	}

	return http.RedirectHandler("/request-account/done", http.StatusSeeOther)
}

// Returns nil if there was no pending invite, so the caller should do something else.
// Returns a non-nil response if there was a pending invite. That response should then be responded
// with and the request should not be handeled further.
func (s *Service) FinishInvite(w http.ResponseWriter, r *http.Request, kthid string) httputil.ToResponse {
	idCookie, _ := r.Cookie("invite")
	if idCookie == nil {
		return nil
	}
	id, err := uuid.Parse(idCookie.Value)
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
		return httputil.BadRequest("This invite has reached its usage limit")
	}
	person, err := kthldap.Lookup(r.Context(), kthid)
	if err != nil {
		return err
	}
	if person == nil {
		slog.Error("Could not find user in ldap", "kthid", kthid, "invite id", id)
		return errors.New("Could not find user in ldap")
	}
	tx, err := s.DB.Begin(r.Context())
	if err != nil {
		return err
	}
	defer tx.Rollback(r.Context())
	if err := tx.CreateUser(r.Context(), database.CreateUserParams{
		Kthid:      kthid,
		UgKthid:    person.UGKTHID,
		Email:      kthid + "@kth.se",
		FirstName:  person.FirstName,
		FamilyName: person.FamilyName,
	}); err != nil {
		return err
	}
	if err := tx.IncrementInviteUses(r.Context(), id); err != nil {
		return err
	}
	if err := tx.Commit(r.Context()); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: "invite", MaxAge: -1})
	slog.Info("User invite link used", "kthid", kthid, "invite-id", inv.ID)
	return s.LoginUser(r.Context(), kthid, true)
}
