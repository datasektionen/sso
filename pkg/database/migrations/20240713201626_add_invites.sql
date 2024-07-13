-- +goose Up
-- +goose StatementBegin
create table invites (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    created_at timestamp not null default now(),
    expires_at timestamp not null,
    max_uses int null,
    current_uses int not null default 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table invites;
-- +goose StatementEnd
