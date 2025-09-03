-- +goose Up
-- +goose StatementBegin
alter table oidc_clients
add column hive_system_id text not null default '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table oidc_clients
drop column hive_system_id;
-- +goose StatementEnd
