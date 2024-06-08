-- +goose Up
-- +goose StatementBegin
create extension "pgcrypto"; -- needed for gen_random_bytes() (and for stonks of course)

create table if not exists users (
    kthid text primary key,
    ug_kthid text not null,
    email text not null,
    first_name text not null,
    surname text not null,
    year_tag text not null,

    member_to date null,

    webauthn_id bytea not null default gen_random_bytes(64)
);

create table if not exists sessions (
    id uuid primary key default gen_random_uuid(),
    kthid text not null,
    last_used_at timestamp not null default now(),

    foreign key (kthid) references users (kthid)
);

create table if not exists passkeys (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    kthid text not null,
    data text not null,

    foreign key (kthid) references users (kthid)
);

create table if not exists legacyapi_tokens (
    id uuid primary key default gen_random_uuid(),
    kthid text not null unique,
    last_used_at timestamp default now(),

    foreign key (kthid) references users (kthid)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
drop table sessions;
drop table passkeys;
drop table legacyapi_tokens;
drop extension "pgcrypto";
-- +goose StatementEnd
