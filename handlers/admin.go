package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/email"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/pkg/kthldap"
	"github.com/datasektionen/sso/pkg/pls"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xuri/excelize/v2"
)

func authAdmin(s *service.Service, h http.Handler) http.Handler {
	return httputil.Route(s, func(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {

		kthid, err := s.GetLoggedInKTHID(r)
		if err != nil {
			return err
		}
		if kthid == "" {
			s.RedirectToLogin(w, r, r.URL.Path)
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

		h.ServeHTTP(w, r)

		if r.Method != http.MethodGet {
			args := []any{"kthid", kthid, "method", r.Method, "path", r.URL.Path}
			if id := r.PathValue("id"); id != "" {
				args = append(args, "id", id)
			} else if r.Form != nil && r.Form.Has("id") {
				args = append(args, "id", r.Form.Get("id"))
			}
			slog.InfoContext(r.Context(), "Admin action taken", args...)
		}

		return nil
	})
}

func admin(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	return templates.AdminPage()
}

func membersPage(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	uploadTime, err := s.DB.GetLastSheetUploadTime(r.Context())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return templates.Members(uploadTime.Time)
}

func adminUsersForm(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	search := r.FormValue("search")
	offsetStr := r.FormValue("offset")
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	year := r.FormValue("year")
	if err != nil && offsetStr != "" {
		return httputil.BadRequest("Invalid int for offset")
	}
	us, err := s.DB.ListUsers(r.Context(), database.ListUsersParams{
		Search: search,
		Limit:  21,
		Offset: int32(offset),
		Year:   year,
	})
	users := service.DBUsersToModel(us)
	if err != nil {
		return err
	}
	more := false
	if len(users) == 21 {
		users = users[0:20:20]
		more = true
	}
	years, err := s.DB.GetAllYears(r.Context())
	if err != nil {
		return err
	}
	return templates.MemberList(users, search, int(offset), more, years, year)
}

func invites(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	invs, err := s.DB.ListInvites(r.Context())
	if err != nil {
		return err
	}
	return templates.Invites(invs)
}

func invite(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	inv, err := s.DB.GetInvite(r.Context(), id)
	if err != nil {
		return httputil.BadRequest("No such invite")
	}
	return templates.Invite(inv)
}

func createInvite(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
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
	inv, err := s.DB.CreateInvite(r.Context(), database.CreateInviteParams{
		Name:      name,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
		MaxUses:   pgtype.Int4{Int32: int32(maxUses), Valid: maxUsesStr != ""},
	})
	if err != nil {
		return err
	}
	return templates.Invite(inv)
}

func deleteInvite(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid id")
	}
	if err := s.DB.DeleteInvite(r.Context(), id); err == pgx.ErrNoRows {
		return httputil.BadRequest("No such invite")
	} else if err != nil {
		return err
	}
	return nil
}

func editInviteForm(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid id")
	}
	invite, err := s.DB.GetInvite(r.Context(), id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such invite")
	}
	return templates.EditInvite(invite)
}

func updateInvite(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
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
	inv, err := s.DB.UpdateInvite(r.Context(), database.UpdateInviteParams{
		ID:        id,
		Name:      name,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
		MaxUses:   pgtype.Int4{Int32: int32(maxUses), Valid: maxUsesStr != ""},
	})
	if err != nil {
		return err
	}
	return templates.Invite(inv)
}

var memberSheet struct {
	// This is locked when retrieving or assigning the reader channel.
	mu sync.Mutex
	// The post handler will instantiate this channel and finish the response.
	// It will then wait for an "event channel" to be sent on this channel
	// until it begins processing the uploaded sheet and then continuously send
	// events from the sheet handling on the "event channel". After processing,
	// this channel will be closed and reassigned to nil.
	//
	// The get handler will send an "event channel" on this channel, read
	// events from that and send them along to the client with SSE.
	reader chan<- chan<- sheetEvent
}

type sheetEvent struct {
	name      string
	component templ.Component
}

func uploadSheet(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	memberSheet.mu.Lock()
	defer memberSheet.mu.Unlock()
	if memberSheet.reader != nil {
		return httputil.BadRequest("Membership sheet upload currently in progress")
	}
	reader := make(chan chan<- sheetEvent)
	memberSheet.reader = reader
	sheet, _, err := r.FormFile("sheet")
	if err != nil {
		return httputil.BadRequest("")
	}
	data, err := io.ReadAll(sheet)
	if err != nil {
		return err
	}
	ctx := context.WithoutCancel(r.Context())
	go func() {
		defer func() {
			close(reader)
			memberSheet.mu.Lock()
			memberSheet.reader = nil
			memberSheet.mu.Unlock()
		}()

		var events chan<- sheetEvent
		t := time.NewTimer(time.Second * 10)
		select {
		case e := <-reader:
			events = e
			t.Stop()
		case <-t.C:
			return
		}

		defer func() {
			events <- sheetEvent{"message", templates.UploadMessage("Done!", false)}
		}()

		const (
			sheetDateCol    = "Giltig till"
			sheetEmailCol   = "E-postadress"
			sheetChapterCol = "Grupp"
		)

		sheet, err := excelize.OpenReader(bytes.NewBuffer(data))
		if err != nil {
			events <- sheetEvent{"message", templates.UploadMessage("Could not parse sheet: "+err.Error(), true)}
			return
		}
		if sheet.SheetCount < 1 {
			events <- sheetEvent{"message", templates.UploadMessage("No sheets found in the provided file", true)}
			return
		}
		rows, err := sheet.GetRows(sheet.GetSheetName(0))
		if err != nil {
			events <- sheetEvent{"message", templates.UploadMessage("No sheets found in the provided file: "+err.Error(), true)}
			return
		}
		if len(rows) == 0 {
			events <- sheetEvent{"message", templates.UploadMessage("Header (first) row not found", true)}
			return
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
			events <- sheetEvent{"message", templates.UploadMessage("Could not find a column for dates", true)}
			return
		}
		if emailCol == -1 {
			events <- sheetEvent{"message", templates.UploadMessage("Could not find a column for emails", true)}
			return
		}
		if chapterCol == -1 {
			events <- sheetEvent{"message", templates.UploadMessage("Could not find a column for chapters", true)}
			return
		}

		for i, columns := range rows[1:] {
			if len(columns) == 0 {
				continue
			}
			if dateCol >= len(columns) || emailCol >= len(columns) || chapterCol >= len(columns) {
				events <- sheetEvent{"message", templates.UploadMessage(fmt.Sprintf(
					"Some column (with index %d, %d or %d) not found on row '%s' with length %d",
					dateCol,
					emailCol,
					chapterCol,
					strings.Join(columns, ","),
					len(columns),
				), true)}
				continue
			}
			date := columns[dateCol]
			email := columns[emailCol]
			chapter := columns[chapterCol]

			kthid, found := strings.CutSuffix(email, "@kth.se")
			if !found {
				events <- sheetEvent{"message", templates.UploadMessage(fmt.Sprintf(
					"Cannot get kth id from row '%s'. Complain to THS that they should have kth-ids in their membership sheet :)",
					strings.Join(columns, ","),
				), true)}
				continue
			}

			if !strings.Contains(chapter, "Datasektionen") {
				if err := s.DB.UserSetMemberTo(ctx, database.UserSetMemberToParams{
					Kthid:    kthid,
					MemberTo: pgtype.Date{Time: time.Now(), Valid: true},
				}); err != nil {
					events <- sheetEvent{"message", templates.UploadMessage(fmt.Sprintf(
						"Could not end membership for user '%s': %v",
						kthid,
						err,
					), true)}
				}
				continue
			}

			memberTo, err := time.Parse(time.DateOnly, date)
			if err != nil {
				events <- sheetEvent{"message", templates.UploadMessage(fmt.Sprintf(
					"Invalid date '%s' for user '%s': %v",
					date,
					kthid,
					err,
				), true)}
			}

			if err := s.DB.Tx(ctx, func(db *database.Queries) error {
				_, err := db.GetUser(ctx, kthid)
				if err == pgx.ErrNoRows {
					person, err := kthldap.Lookup(ctx, kthid)
					if err != nil {
						return err
					}
					if person == nil {
						events <- sheetEvent{"message", templates.UploadMessage(fmt.Sprintf(
							"Could not find user with kthid '%s' in KTH's ldap",
							kthid,
						), true)}
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
				events <- sheetEvent{"message", templates.UploadMessage(fmt.Sprintf(
					"Could not update user '%s' in database: %v",
					kthid,
					err,
				), true)}
			}
			events <- sheetEvent{"progress", templates.UploadProgress(float64(i) / float64(len(rows)-1))}
		}
		if err := s.DB.MarkSheetUploadedNow(ctx); err != nil {

			events <- sheetEvent{"message", templates.UploadMessage("Could not mark last uploaded date so you will forever think the fresh data is old 😈", false)}
		}
		events <- sheetEvent{"progress", templates.UploadProgress(1)}
	}()
	return templates.UploadStatus(true)
}

func processSheet(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	memberSheet.mu.Lock()
	reader := memberSheet.reader
	memberSheet.mu.Unlock()
	if r == nil {
		return httputil.BadRequest("No membership sheet upload waiting to get started")
	}

	ch := make(chan sheetEvent)
	reader <- ch

	// Yes, server-sent events actually are that easy
	w.Header().Set("Content-Type", "text/event-stream")
	flusher, canFlush := w.(interface{ Flush() })
	for event := range ch {
		_, _ = w.Write([]byte("event: " + event.name + "\n"))
		var buf bytes.Buffer
		if event.component != nil {
			_ = event.component.Render(r.Context(), &buf)
		}
		for _, line := range strings.Split(buf.String(), "\n") {
			_, _ = w.Write([]byte("data: " + line + "\n"))
		}
		_, _ = w.Write([]byte("\n"))
		if canFlush {
			flusher.Flush()
		}
	}

	return nil
}

func oidcClients(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	clients, err := s.DB.ListClients(r.Context())
	if err != nil {
		return err
	}
	return templates.OidcClients(clients)
}

var oidcClientIDRegexp = regexp.MustCompile(`^[a-z\-0-1]+$`)

func createOIDCClient(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id := r.FormValue("id")
	if !oidcClientIDRegexp.Match([]byte(id)) {
		return httputil.BadRequest("ID must match " + oidcClientIDRegexp.String())
	}

	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return err
	}

	h := sha256.New()
	h.Write(secret[:])
	secretHash := h.Sum(nil)

	client, err := s.DB.CreateClient(r.Context(), database.CreateClientParams{
		ID:         id,
		SecretHash: secretHash,
	})
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23505" /* unique_violation */ {
			return httputil.BadRequest("Client ID already taken")
		}
		return err
	}
	return templates.OidcClient(client, secret[:])
}

func deleteOIDCClient(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id := r.PathValue("id")

	if err := s.DB.DeleteClient(r.Context(), id); err == pgx.ErrNoRows {
		return httputil.BadRequest("No such client")
	} else if err != nil {
		return err
	}
	return nil
}

func addRedirectURI(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id := r.PathValue("id")
	newURI := r.FormValue("redirect-uri")
	if newURI == "" {
		return httputil.BadRequest("Missing uri")
	}

	client, err := s.DB.GetClient(r.Context(), id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such client")
	} else if err != nil {
		return err
	}

	client.RedirectUris = append(client.RedirectUris, newURI)

	if _, err := s.DB.UpdateClient(
		r.Context(),
		database.UpdateClientParams{
			ID:           client.ID,
			RedirectUris: client.RedirectUris,
		},
	); err != nil {
		return err
	}

	return templates.RedirectURI(id, newURI)
}

func removeRedirectURI(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id := r.PathValue("id")
	uri := r.PathValue("uri")

	client, err := s.DB.GetClient(r.Context(), id)
	if err == pgx.ErrNoRows {
		return httputil.BadRequest("No such client")
	} else if err != nil {
		return err
	}

	client.RedirectUris = slices.DeleteFunc(client.RedirectUris, func(u string) bool { return u == uri })

	if _, err := s.DB.UpdateClient(
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

func accountRequests(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	requests, err := s.DB.ListAccountRequests(r.Context())
	if err != nil {
		return err
	}

	requests = slices.DeleteFunc(requests, func(req database.AccountRequest) bool { return !req.Kthid.Valid })

	return templates.AccountRequests(requests)
}

func denyAccountRequest(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}
	req, err := s.DB.DeleteAccountRequest(r.Context(), id)
	if err != nil {
		return err
	}

	if kthid := req.Kthid.String; r.URL.Query().Has("email") && req.Kthid.Valid && kthid != "" {
		if err := email.Send(r.Context(), kthid+"@kth.se", "Datasektionen account request denied", "<p>Your Datasektionen account request has been denied.</p>"); err != nil {
			slog.Error("Could not send email", "error", err)
			return "Denied, but could not send email!"
		}
		return "Denied ❌ and sent email!"
	}

	return "Denied ❌"
}

func approveAccountRequest(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httputil.BadRequest("Invalid uuid")
	}

	tx, err := s.DB.Begin(r.Context())
	if err != nil {
		return err
	}
	defer tx.Rollback(r.Context())

	accountRequest, err := tx.DeleteAccountRequest(r.Context(), id)
	if err != nil {
		return err
	}

	kthid := accountRequest.Kthid.String
	if kthid == "" {
		return httputil.BadRequest("No KTH ID - approval not implemented for that")
	}

	person, err := kthldap.Lookup(r.Context(), accountRequest.Kthid.String)
	if err != nil {
		return err
	}
	if person == nil {
		return fmt.Errorf("Could not find user with kthid '%s' in KTH's ldap", kthid)
	}
	emailAddress := person.KTHID + "@kth.se"
	if err := tx.CreateUser(r.Context(), database.CreateUserParams{
		Kthid:      kthid,
		UgKthid:    person.UGKTHID,
		Email:      emailAddress,
		FirstName:  person.FirstName,
		FamilyName: person.FamilyName,
	}); err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23505" /* unique_violation */ {
			return httputil.BadRequest("User already exists")
		}
		return err
	}

	if err := tx.Commit(r.Context()); err != nil {
		return err
	}

	if err := email.Send(r.Context(), emailAddress, "Datasektionen account request approved", strings.TrimSpace(fmt.Sprintf(`
Hello %s, your Datasektionen account request has been approved!

You can go to [sso.datasektionen.se](https://sso.datasektionen.se/) to log in and see your account, or simply go directly to a system you want to log in to.
	`, person.FirstName))); err != nil {
		slog.Error("Could not send email", "error", err)
		return "Approved, but could not send email!"
	}

	return "Approved ✅"
}
