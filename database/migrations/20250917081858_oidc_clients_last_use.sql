-- +goose Up
-- +goose StatementBegin
alter table oidc_clients
add column last_used_at timestamp null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table oidc_clients
drop column last_used_at;
-- +goose StatementEnd
