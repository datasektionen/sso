-- +goose Up
-- +goose StatementBegin
alter table users add column nfc_id text not null default ''::text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users drop column nfc_id;
-- +goose StatementEnd
