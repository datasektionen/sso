package main

import (
	"context"
	"crypto/rand"
	"encoding/json"

	"database/sql"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type User struct {
	KTHID      string `db:"kthid"`
	WebAuthnID []byte `db:"webauthn_id"`
	Passkeys   []Passkey
}

type Passkey struct {
	ID   uuid.UUID
	Name string
	Cred webauthn.Credential
}

type WebAuthnUser struct{ *User }

func (u WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	res := make([]webauthn.Credential, len(u.User.Passkeys))
	for i, cred := range u.User.Passkeys {
		res[i] = cred.Cred
	}
	return res
}

func (u WebAuthnUser) WebAuthnDisplayName() string {
	// TODO: use full name
	return u.KTHID
}

func (u WebAuthnUser) WebAuthnID() []byte {
	return u.User.WebAuthnID[:]
}

func (u WebAuthnUser) WebAuthnIcon() string {
	// NOTE: No loger in the spec, library recommends empty string
	return ""
}

func (u WebAuthnUser) WebAuthnName() string {
	return u.KTHID
}

type DB struct {
	db *sqlx.DB
}

func NewDB(uri string) (DB, error) {
	db, err := sqlx.Connect("pgx", uri)
	if err != nil {
		return DB{}, err
	}
	_, err = db.Exec(`--sql
		-- drop schema public cascade; create schema public;
		create table if not exists users (
			kthid text primary key,
			webauthn_id bytea not null
		);
		create table if not exists passkeys (
			id uuid primary key default gen_random_uuid(),
			name text not null default to_char(now(), 'yyyy-mm-dd'),
			kthid text not null,
			data text not null,

			foreign key (kthid) references users (kthid)
		);
		create table if not exists sessions (
			id uuid primary key default gen_random_uuid(),
			kthid text not null,

			foreign key (kthid) references users (kthid)
		);
	`)
	if err != nil {
		return DB{}, err
	}
	return DB{db}, nil
}

func (db DB) CreateSession(ctx context.Context, kthid string) (uuid.UUID, error) {
	var sessionID uuid.UUID
	if err := db.db.QueryRowxContext(ctx, `--sql
		insert into sessions (kthid) values ($1) returning id
	`, kthid).Scan(&sessionID); err != nil {
		return uuid.Nil, err
	}
	return sessionID, nil
}

func (db DB) GetSession(sessionID string) (string, error) {
	var kthid string
	if err := db.db.QueryRowx(`--sql
		select kthid from sessions where id = $1
	`, sessionID).Scan(&kthid); err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return kthid, nil
}

func (db DB) RemoveSession(ctx context.Context, sessionID string) error {
	_, err := db.db.ExecContext(ctx, `--sql
		delete from sessions where id = $1
	`, sessionID)
	return err
}

func (db DB) CreateUser(ctx context.Context, kthid string) error {
	var webAuthnID [64]byte
	if _, err := rand.Read(webAuthnID[:]); err != nil {
		return err
	}
	if _, err := db.db.ExecContext(ctx, `--sql
		insert into users (kthid, webauthn_id) values ($1, $2)
	`, kthid, webAuthnID[:]); err != nil {
		return err
	}
	return nil
}

func (db DB) AddPasskey(ctx context.Context, kthid string, name string, cred webauthn.Credential) error {
	credJSON, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	_, err = db.db.ExecContext(ctx, `--sql
		insert into passkeys (kthid, name, data)
		values ($1, $2, $3)
	`, kthid, name, string(credJSON))
	return err
}

func (db DB) RemovePasskey(ctx context.Context, kthid string, passkeyID uuid.UUID) error {
	res, err := db.db.ExecContext(ctx, `--sql
		delete from passkeys
		where kthid = $1 and id = $2
	`, kthid, passkeyID)
	if r, err := res.RowsAffected(); err != nil {
		return err
	} else if r == 0 {
		return BadRequest("No such passkey id")
	}
	return err
}

func (db DB) GetUser(ctx context.Context, kthid string) (*User, error) {
	rows, err := db.db.Queryx(`--sql
		select u.*, wc.id as wc_id, wc.name as wc_name, wc.data as wc_data
		from users u
		left join passkeys wc
			on wc.kthid = u.kthid
		where u.kthid = $1
	`, kthid)
	if err != nil {
		return nil, err
	}
	var user User
	for rows.Next() {
		row := make(map[string]any)
		rows.MapScan(row)
		user.KTHID = row["kthid"].(string)
		user.WebAuthnID = row["webauthn_id"].([]byte)
		if row["wc_id"] == nil {
			break
		}
		wcID, err := uuid.Parse(row["wc_id"].(string))
		if err != nil {
			return nil, err
		}
		wcName := row["wc_name"].(string)
		var cred webauthn.Credential
		if err := json.Unmarshal([]byte(row["wc_data"].(string)), &cred); err != nil {
			return nil, err
		}
		user.Passkeys = append(user.Passkeys, Passkey{
			ID:   wcID,
			Name: wcName,
			Cred: cred,
		})
	}
	if user.KTHID == "" {
		return nil, nil
	}

	return &user, nil
}
