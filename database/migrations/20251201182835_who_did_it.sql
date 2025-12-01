-- +goose Up
-- +goose StatementBegin
alter table last_membership_sheet
add column uploaded_by text not null default 'mrbean';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table last_membership_sheet
drop column uploaded_by;
-- +goose StatementEnd
