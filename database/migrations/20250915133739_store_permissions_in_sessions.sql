-- +goose Up
-- +goose StatementBegin
delete from sessions;
alter table sessions
add column permissions json not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table sessions
drop column permissions;
-- +goose StatementEnd
