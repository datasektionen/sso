-- +goose Up
-- +goose StatementBegin
create table email_login (
    kthid text not null,
    code text not null,
    creation_time timestamp with time zone not null default now(),
    attempts int not null default 0,
    primary key (kthid, creation_time)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table email_login;
-- +goose StatementEnd
