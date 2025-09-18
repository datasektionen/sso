-- +goose Up
-- +goose StatementBegin
alter table passkeys
alter column data type json using data::json;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table passkeys
alter column data type text;
-- +goose StatementEnd
