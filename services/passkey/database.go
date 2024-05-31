package passkey

import (
	"context"
	"encoding/json"

	"github.com/datasektionen/logout/pkg/httputil"
	"github.com/datasektionen/logout/services/passkey/export"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

func (s *service) migrateDB() error {
	_, err := s.db.Exec(`--sql
		create table if not exists passkeys (
			id uuid primary key default gen_random_uuid(),
			name text not null,
			kthid text not null,
			data text not null,

			foreign key (kthid) references users (kthid)
		);
	`)
	return err
}

func (s *service) AddPasskey(ctx context.Context, kthid string, name string, cred webauthn.Credential) error {
	credJSON, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `--sql
		insert into passkeys (kthid, name, data)
		values ($1, $2, $3)
	`, kthid, name, string(credJSON))
	return err
}

func (s *service) RemovePasskey(ctx context.Context, kthid string, passkeyID uuid.UUID) error {
	res, err := s.db.ExecContext(ctx, `--sql
		delete from passkeys
		where kthid = $1 and id = $2
	`, kthid, passkeyID)
	if r, err := res.RowsAffected(); err != nil {
		return err
	} else if r == 0 {
		return httputil.BadRequest("No such passkey id")
	}
	return err
}

func (s *service) GetPasskeysForUser(ctx context.Context, kthid string) ([]export.Passkey, error) {
	res, err := s.db.QueryxContext(ctx, `--sql
		select id, name, data
		from passkeys
		where kthid = $1
	`, kthid)
	if err != nil {
		return nil, err
	}
	var passkeys []export.Passkey
	for res.Next() {
		var p struct {
			ID   uuid.UUID
			Name string
			Data string
		}
		if err := res.StructScan(&p); err != nil {
			return nil, err
		}
		pk := export.Passkey{ID: p.ID, Name: p.Name}
		if err := json.Unmarshal([]byte(p.Data), &pk.Cred); err != nil {
			return nil, err
		}
		passkeys = append(passkeys, pk)
	}
	return passkeys, nil
}
