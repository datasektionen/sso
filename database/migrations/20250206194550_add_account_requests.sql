-- +goose Up
-- +goose StatementBegin
create table account_requests (
    id uuid primary key default gen_random_uuid(),
    created_at timestamp default now(),
    reference text not null,
    reason text not null,
    year_tag text not null,

    kthid text null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table account_requests;
-- +goose StatementEnd
