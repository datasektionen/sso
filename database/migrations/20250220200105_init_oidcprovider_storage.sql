-- +goose Up
-- +goose StatementBegin
create table oidcprovider_access_tokens (
    id uuid primary key default gen_random_uuid(),
    kthid text not null,
    scopes text[] not null,

    foreign key (kthid) references users (kthid)
);

create table oidcprovider_auth_requests (
    id uuid primary key default gen_random_uuid(),
    auth_code text not null,
    kthid text not null,

    data json not null,

    foreign key (kthid) references users (kthid)
);

create index oidcprovider_auth_requests_auth_code_idx
on oidcprovider_auth_requests (auth_code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table oidcprovider_auth_requests;
drop table oidcprovider_access_tokens;
-- +goose StatementEnd
