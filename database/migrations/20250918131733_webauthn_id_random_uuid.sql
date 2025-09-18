-- +goose Up
-- +goose StatementBegin
drop table webauthn_session_data;
create table webauthn_session_data (
    id uuid primary key default gen_random_uuid(),
    kthid text not null,
    created_at timestamp default now(),
    data json not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table webauthn_session_data;
create table webauthn_session_data (
    kthid text primary key,
    data json not null
);
-- +goose StatementEnd
