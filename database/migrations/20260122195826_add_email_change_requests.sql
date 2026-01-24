-- +goose Up
-- +goose StatementBegin
create table email_change_requests (
    kthid text primary key references users(kthid),
    new_email text not null,
    code text not null,
    created_at timestamp not null default now(),
    attempts int not null default 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table email_change_requests;
-- +goose StatementEnd
