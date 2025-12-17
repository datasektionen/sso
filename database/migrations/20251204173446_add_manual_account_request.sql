-- +goose Up
-- +goose StatementBegin
alter table account_requests
add column ug_kthid text;

alter table account_requests
add column first_name text;

alter table account_requests
add column family_name text;

alter table account_requests
add column email text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table account_requests
drop column ug_kthid;

alter table account_requests
drop column first_name;

alter table account_requests
drop column family_name;

alter table account_requests
drop column email;
-- +goose StatementEnd
