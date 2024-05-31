package legacyapi

import (
	"context"

	"github.com/google/uuid"
)

func (s *service) migrateDB(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `--sql
		create table if not exists legacyapi_tokens (
			id uuid primary key default gen_random_uuid(),
			kthid text not null unique,
			last_used_at timestamp default now(),

			foreign key (kthid) references users (kthid)
		);
	`)
	return err
}

func (s *service) createToken(ctx context.Context, kthid string) (uuid.UUID, error) {
	var id uuid.UUID
	if err := s.db.QueryRowContext(ctx, `--sql
		insert into legacyapi_tokens (kthid)
		values ($1)
		on conflict (kthid)
		do update
		set last_used_at = now()
		returning id
	`, kthid).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *service) getToken(ctx context.Context, token uuid.UUID) (string, error) {
	var kthid string
	if err := s.db.QueryRowContext(ctx, `--sql
		update legacyapi_tokens
		set last_used_at = now()
		where id = $1 and last_used_at > now() - interval '8 hours'
		returning kthid
	`, token).Scan(&kthid); err != nil {
		return "", err
	}
	return kthid, nil
}

func (s *service) deleteToken(ctx context.Context, kthid string) error {
	_, err := s.db.ExecContext(ctx, `--sql
		delete from legacyapi_tokens
		where kthid = $1
	`, kthid)
	return err
}
