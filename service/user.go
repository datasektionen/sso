package service

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/datasektionen/logout/database"
	"github.com/datasektionen/logout/models"
	"github.com/datasektionen/logout/pkg/auth"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/pkg/kthldap"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (s *Service) UserSetNameChangeRequest(ctx context.Context, kthid string, firstName string, familyName string) (models.User, error) {
	user, err := s.DB.UserSetNameChangeRequest(ctx, database.UserSetNameChangeRequestParams{
		Kthid:                   kthid,
		FirstNameChangeRequest:  firstName,
		FamilyNameChangeRequest: familyName,
	})
	if err != nil {
		return models.User{}, err
	}
	return dbUserToModel(user), nil
}

func (s *Service) LoginUser(ctx context.Context, kthid string) httputil.ToResponse {
	sessionID, err := s.DB.CreateSession(ctx, kthid)
	if err != nil {
		return err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, auth.SessionCookie(sessionID.String()))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func (s *Service) GetLoggedInKTHID(r *http.Request) (string, error) {
	sessionCookie, _ := r.Cookie(auth.SessionCookieName)
	if sessionCookie == nil {
		return "", nil
	}
	sessionID, err := uuid.Parse(sessionCookie.Value)
	if err != nil {
		return "", nil
	}
	session, err := s.DB.GetSession(r.Context(), sessionID)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return session, nil
}

func (s *Service) GetLoggedInUser(r *http.Request) (*models.User, error) {
	kthid, err := s.GetLoggedInKTHID(r)
	if err != nil {
		return nil, err
	}
	if kthid == "" {
		return nil, nil
	}
	return s.GetUser(r.Context(), kthid)
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
	http.Redirect(w, r, "/?"+url.Values{"next-url": []string{nextURL}}.Encode(), http.StatusSeeOther)
}

func (s *Service) FinishInvite(w http.ResponseWriter, r *http.Request, kthid string) (bool, httputil.ToResponse) {
	idCookie, _ := r.Cookie("invite")
	if idCookie == nil {
		return false, nil
	}
	id, err := uuid.Parse(idCookie.Value)
	if err != nil {
		return true, httputil.BadRequest("Invalid uuid")
	}
	inv, err := s.DB.GetInvite(r.Context(), id)
	if err == pgx.ErrNoRows {
		return true, httputil.BadRequest("No such invite")
	} else if err != nil {
		return true, err
	}
	if time.Now().After(inv.ExpiresAt.Time) {
		return true, httputil.BadRequest("Invite expired")
	}
	if inv.MaxUses.Valid && inv.CurrentUses >= inv.MaxUses.Int32 {
		return true, httputil.BadRequest("This invite has reached its usage limit")
	}
	person, err := kthldap.Lookup(r.Context(), kthid)
	if err != nil {
		return true, err
	}
	if person == nil {
		slog.Error("Could not find user in ldap", "kthid", kthid, "invite id", id)
		return true, errors.New("Could not find user in ldap")
	}
	tx, err := s.DB.Begin(r.Context())
	if err != nil {
		return true, err
	}
	defer tx.Rollback(r.Context())
	if err := tx.CreateUser(r.Context(), database.CreateUserParams{
		Kthid:      kthid,
		UgKthid:    person.UGKTHID,
		Email:      kthid + "@kth.se",
		FirstName:  person.FirstName,
		FamilyName: person.FamilyName,
	}); err != nil {
		return true, err
	}
	if err := tx.IncrementInviteUses(r.Context(), id); err != nil {
		return true, err
	}
	if err := tx.Commit(r.Context()); err != nil {
		return true, err
	}
	http.SetCookie(w, &http.Cookie{Name: "invite", MaxAge: -1})
	slog.Info("User invite link used", "kthid", kthid, "invite-id", inv.ID)
	return true, s.LoginUser(r.Context(), kthid)
}
