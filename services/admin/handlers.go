package admin

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/datasektionen/logout/pkg/database"
	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/pkg/kthldap"
	"github.com/datasektionen/logout/pkg/pls"
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
			http.Redirect(w, r, "/", http.StatusSeeOther)
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
	resp := make([]struct {
		ID           string   `json:"id"`
		RedirectURIs []string `json:"redirect_uris"`
	}, len(clients))
	for i, client := range clients {
		resp[i].ID = base64.URLEncoding.EncodeToString(client.ID)
		resp[i].RedirectURIs = client.RedirectUris
	}
	return httputil.JSON(resp)
}

func (s *service) createOIDCClient(w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	var body struct {
		RedirectURIs []string `json:"redirect_uris"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return httputil.BadRequest("Invalid json in body")
	}

	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return err
	}

	h := sha256.New()
	h.Write(secret[:])
	id := h.Sum(nil)

	if err := s.db.CreateClient(r.Context(), database.CreateClientParams{
		ID:           id,
		RedirectUris: body.RedirectURIs,
	}); err != nil {
		return err
	}
	return httputil.JSON(map[string]any{
		"id":     base64.URLEncoding.EncodeToString(id),
		"secret": base64.URLEncoding.EncodeToString(secret[:]),
	})
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
