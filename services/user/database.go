package user

import (
	"context"
	"crypto/rand"
	"database/sql"

	"github.com/datasektionen/logout/services/user/export"
	"github.com/google/uuid"
)

func (s *service) migrateDB() error {
	_, err := s.db.Exec(`--sql
		-- drop schema public cascade; create schema public;
		create table if not exists users (
			kthid text primary key,
			webauthn_id bytea not null
		);
		create table if not exists sessions (
			id uuid primary key default gen_random_uuid(),
			kthid text not null,
			last_used_at timestamp not null default now(),

			foreign key (kthid) references users (kthid)
		);
    `)
	return err
}

func (s *service) CreateSession(ctx context.Context, kthid string) (uuid.UUID, error) {
	var sessionID uuid.UUID
	if err := s.db.QueryRowxContext(ctx, `--sql
		insert into sessions (kthid) values ($1) returning id
	`, kthid).Scan(&sessionID); err != nil {
		return uuid.Nil, err
	}
	return sessionID, nil
}

func (s *service) GetSession(sessionID string) (string, error) {
	var kthid string
	if err := s.db.QueryRowx(`--sql
		update sessions
		set last_used_at = now()
		where id = $1
		and last_used_at > now() - interval '8 hours'
		returning kthid
	`, sessionID).Scan(&kthid); err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return kthid, nil
}

func (s *service) RemoveSession(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `--sql
		delete from sessions where id = $1
	`, sessionID)
	return err
}

func (s *service) CreateUser(ctx context.Context, kthid string) error {
	var webAuthnID [64]byte
	if _, err := rand.Read(webAuthnID[:]); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `--sql
		insert into users (kthid, webauthn_id) values ($1, $2)
	`, kthid, webAuthnID[:]); err != nil {
		return err
	}
	return nil
}

func (s *service) GetUser(ctx context.Context, kthid string) (*export.User, error) {
	var user export.User
	if err := s.db.GetContext(ctx, &user, `--sql
		select *
		from users
		where kthid = $1
	`, kthid); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}
