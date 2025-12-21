-- +goose Up
-- +goose StatementBegin
create table email_logins (
    kthid text primary key, -- each user may only have one code "active" at a time
    code text not null,
    created_at timestamp not null default now(),
    attempts int not null default 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table email_logins;
-- +goose StatementEnd
