-- +goose Up
-- +goose StatementBegin
alter table oidc_clients
add column allow_nollan boolean not null default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table oidc_clients
drop column allow_nollan;
-- +goose StatementEnd
