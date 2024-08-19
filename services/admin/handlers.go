package admin

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/pkg/kthldap"
	"github.com/datasektionen/logout/pkg/pls"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/xuri/excelize/v2"
)

func (s *service) auth(h http.Handler) http.Handler {
	return httputil.Route(func(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
		kthid, err := s.user.GetLoggedInKTHID(r)
		if err != nil {
			return err
		}
		if kthid == "" {
			s.user.RedirectToLogin(w, r, "/admin")
			return nil
		}
		perm := "admin-write"
		if r.Method == http.MethodGet {
			perm = "admin-read"
		}
		allowed, err := pls.CheckUser(r.Context(), kthid, perm)
		if err != nil {
			return err
		}
		if !allowed {
			return httputil.Forbidden("Missing admin permission in pls")
		}
		return h
	})
}

func (s *service) admin(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return admin()
}

func (s *service) members(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return members()
}

func (s *service) invites(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	invs, err := s.db.ListInvites(r.Context())
	if err != nil {
		return err
	}
	return invites(invs)
}

func (s *service) invite(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	inv, err := s.db.GetInvite(r.Context(), id)
	if err != nil {
		return httputil.BadRequest("No such invite")
	}
	return invite(inv)
}

func (s *service) createInvite(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	name := r.FormValue("name")
	expiresAt, err := time.Parse(time.DateOnly, r.FormValue("expires-at"))
	if err != nil {
		return httputil.BadRequest("Invalid date for expires at")
	}
	maxUsesStr := r.FormValue("max-uses")
	maxUses, err := strconv.Atoi(maxUsesStr)
	if err != nil && maxUsesStr != "" {
		return httputil.BadRequest("Invalid int for max uses")
	}
	inv, err := s.db.CreateInvite(r.Context(), database.CreateInviteParams{
		Name:      name,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
		MaxUses:   pgtype.Int4{Int32: int32(maxUses), Valid: maxUsesStr != ""},
	})
	if err != nil {
		return err
	}
	return invite(inv)
}

func (s *service) deleteInvite(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid id")
	}
	if err := s.db.DeleteInvite(r.Context(), id); err == pgx.ErrNoRows {
		return httputil.BadRequest("No such invite")
	} else if err != nil {
		return err
	}
	return nil
}

func (s *service) editInviteForm(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid id")
	}
	invite, err := s.db.GetInvite(r.Context(), id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such invite")
	}
	return editInvite(invite)
}

func (s *service) updateInvite(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid id")
	}

	name := r.FormValue("name")
	expiresAt, err := time.Parse(time.DateOnly, r.FormValue("expires-at"))
	if err != nil {
		return httputil.BadRequest("Invalid date for expires at")
	}
	maxUsesStr := r.FormValue("max-uses")
	maxUses, err := strconv.Atoi(maxUsesStr)
	if err != nil && maxUsesStr != "" {
		return httputil.BadRequest("Invalid int for max uses")
	}
	inv, err := s.db.UpdateInvite(r.Context(), database.UpdateInviteParams{
		ID:        id,
		Name:      name,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
		MaxUses:   pgtype.Int4{Int32: int32(maxUses), Valid: maxUsesStr != ""},
	})
	if err != nil {
		return err
	}
	return invite(inv)
}

func (s *service) uploadSheet(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	s.memberSheet.mu.Lock()
	defer s.memberSheet.mu.Unlock()
	if s.memberSheet.inProgress {
		return httputil.BadRequest("Membership sheet upload currently in progress")
	}
	var err error
	s.memberSheet.data, err = io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) processSheet(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	ctx := r.Context()
	s.memberSheet.mu.Lock()
	if s.memberSheet.inProgress {
		return httputil.BadRequest("Membership sheet upload already in progress")
	}
	s.memberSheet.inProgress = true
	s.memberSheet.mu.Unlock()
	defer func() {
		s.memberSheet.mu.Lock()
		s.memberSheet.inProgress = false
		s.memberSheet.data = nil
		s.memberSheet.mu.Unlock()
	}()

	// Yes, server-sent events actually are that easy
	w.Header().Set("Content-Type", "text/event-stream")
	flusher, canFlush := w.(interface{ Flush() })
	event := func(event, data string) {
		w.Write([]byte("event: " + event + "\n"))
		for _, line := range strings.Split(data, "\n") {
			w.Write([]byte("data: " + line + "\n"))
		}
		w.Write([]byte("\n"))
		if canFlush {
			flusher.Flush()
		}
	}

	defer event("done", "")

	const (
		sheetDateCol    = "Giltig till"
		sheetEmailCol   = "E-postadress"
		sheetChapterCol = "Grupp"
	)

	sheet, err := excelize.OpenReader(bytes.NewBuffer(s.memberSheet.data))
	if err != nil {
		event("err", "Could not parse sheet: "+err.Error())
		return nil
	}
	if sheet.SheetCount < 1 {
		event("err", "No sheets found in the provided file")
		return nil
	}
	rows, err := sheet.GetRows(sheet.GetSheetName(0))
	if err != nil {
		event("err", "No sheets found in the provided file: "+err.Error())
		return nil
	}
	if len(rows) == 0 {
		event("err", "Header (first) row not found")
		return nil
	}
	var dateCol, emailCol, chapterCol int = -1, -1, -1
	for i, title := range rows[0] {
		title = strings.TrimSpace(title)
		if title == sheetDateCol {
			dateCol = i
		} else if title == sheetEmailCol {
			emailCol = i
		} else if title == sheetChapterCol {
			chapterCol = i
		}
	}
	if dateCol == -1 {
		event("err", "Could not find a column for dates")
		return nil
	}
	if emailCol == -1 {
		event("err", "Could not find a column for emails")
		return nil
	}
	if chapterCol == -1 {
		event("err", "Could not find a column for chapters")
		return nil
	}

	for i, columns := range rows[1:] {
		if len(columns) == 0 {
			continue
		}
		if dateCol >= len(columns) || emailCol >= len(columns) || chapterCol >= len(columns) {
			event("err", fmt.Sprintf(
				"Some column (with index %d, %d or %d) not found on row '%s' with length %d",
				dateCol,
				emailCol,
				chapterCol,
				strings.Join(columns, ","),
				len(columns),
			))
			continue
		}
		date := columns[dateCol]
		email := columns[emailCol]
		chapter := columns[chapterCol]

		kthid, found := strings.CutSuffix(email, "@kth.se")
		if !found {
			event("err", fmt.Sprintf(
				"Cannot get kth id from row '%s'. Complain to THS that they should have kth-ids in their membership sheet :)",
				strings.Join(columns, ","),
			))
			continue
		}

		if !strings.Contains(chapter, "Datasektionen") {
			if err := s.db.UserSetMemberTo(ctx, database.UserSetMemberToParams{
				Kthid:    kthid,
				MemberTo: pgtype.Date{Time: time.Now(), Valid: true},
			}); err != nil {
				event("err", fmt.Sprintf(
					"Could not end membership for user '%s': %v",
					kthid,
					err,
				))
			}
			continue
		}

		memberTo, err := time.Parse(time.DateOnly, date)
		if err != nil {
			event("err", fmt.Sprintf(
				"Invalid date '%s' for user '%s': %v",
				date,
				kthid,
				err,
			))
		}

		if err := s.db.Tx(ctx, func(db *database.Queries) error {
			_, err := db.GetUser(ctx, kthid)
			if err == pgx.ErrNoRows {
				person, err := kthldap.Lookup(ctx, kthid)
				if err != nil {
					return err
				}
				if person == nil {
					event("err", fmt.Sprintf(
						"Could not find user with kthid '%s' in KTH's ldap",
						kthid,
					))
					return nil
				}
				if err := db.CreateUser(ctx, database.CreateUserParams{
					Kthid:      kthid,
					UgKthid:    person.UGKTHID,
					Email:      email,
					FirstName:  person.FirstName,
					FamilyName: person.FamilyName,
					MemberTo:   pgtype.Date{Valid: true, Time: memberTo},
				}); err != nil {
					return err
				}
				return nil
			}
			if err != nil {
				return err
			}
			return db.UserSetMemberTo(ctx, database.UserSetMemberToParams{
				Kthid:    kthid,
				MemberTo: pgtype.Date{Valid: true, Time: memberTo},
			})
		}); err != nil {
			event("err", fmt.Sprintf(
				"Could not update user '%s' in database: %v",
				kthid,
				err,
			))
		}
		event("progress", fmt.Sprintf("%f", float64(i)/float64(len(rows)-1)))
	}
	event("progress", fmt.Sprintf("%f", 1.0))

	return nil
}

func (s *service) oidcClients(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	clients, err := s.db.ListClients(r.Context())
	if err != nil {
		return err
	}
	return oidcClients(clients)
}

func (s *service) createOIDCClient(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Body must be valid application/x-www-form-urlencoded")
	}

	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return err
	}

	h := sha256.New()
	h.Write(secret[:])
	id := h.Sum(nil)

	client, err := s.db.CreateClient(r.Context(), id)
	if err != nil {
		return err
	}
	return oidcClient(client, secret[:])
}

func (s *service) deleteOIDCClient(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := base64.URLEncoding.DecodeString(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid id")
	}

	if err := s.db.DeleteClient(r.Context(), id); err == pgx.ErrNoRows {
		return httputil.BadRequest("No such client")
	} else if err != nil {
		return err
	}
	return nil
}

func (s *service) addRedirectURI(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := base64.URLEncoding.DecodeString(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid id")
	}
	newURI := r.FormValue("redirect-uri")

	client, err := s.db.GetClient(r.Context(), id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such client")
	} else if err != nil {
		return err
	}

	client.RedirectUris = append(client.RedirectUris, newURI)

	if _, err := s.db.UpdateClient(
		r.Context(),
		database.UpdateClientParams{
			ID:           client.ID,
			RedirectUris: client.RedirectUris,
		},
	); err != nil {
		return err
	}

	return redirectURI(id, newURI)
}

func (s *service) removeRedirectURI(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := base64.URLEncoding.DecodeString(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid id")
	}
	uri := r.PathValue("uri")

	client, err := s.db.GetClient(r.Context(), id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such client")
	} else if err != nil {
		return err
	}

	client.RedirectUris = slices.DeleteFunc(client.RedirectUris, func(u string) bool { return u == uri })

	if _, err := s.db.UpdateClient(
		r.Context(),
		database.UpdateClientParams{
			ID:           client.ID,
			RedirectUris: client.RedirectUris,
		},
	); err != nil {
		return err
	}

	return nil
}
