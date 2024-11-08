-- +goose Up
-- +goose StatementBegin
alter table users
add column first_name_change_request text not null default '', -- '' when no current change request
add column family_name_change_request text not null default ''; -- ---||---
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
drop column first_name_change_request,
drop column family_name_change_request;
-- +goose StatementEnd
