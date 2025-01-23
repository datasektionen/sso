-- +goose Up
-- +goose StatementBegin
create table last_membership_sheet (
    unique_marker text primary key check (unique_marker = 'unique_marker') default 'unique_marker',
    uploaded_at timestamp not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table last_membership_sheet;
-- +goose StatementEnd
