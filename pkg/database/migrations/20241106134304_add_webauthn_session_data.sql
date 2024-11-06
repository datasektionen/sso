-- +goose Up
-- +goose StatementBegin
create table webauthn_session_data (
    kthid text primary key,
    data json not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table webauthn_session_data;
-- +goose StatementEnd
