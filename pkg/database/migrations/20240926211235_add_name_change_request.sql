-- +goose Up
-- +goose StatementBegin
alter table users
add column first_name_change_request text default null,
add column family_name_change_request text default null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
drop column first_name_change_request,
drop column family_name_change_request;
-- +goose StatementEnd
