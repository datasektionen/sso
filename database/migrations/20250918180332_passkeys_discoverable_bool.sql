-- +goose Up
-- +goose StatementBegin
alter table passkeys
add column discoverable bool not null default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table passkeys
drop column discoverable;
-- +goose StatementEnd
