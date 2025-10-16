-- +goose Up
-- +goose StatementBegin
alter table sessions
drop constraint sessions_kthid_fkey;

alter table sessions
alter column kthid drop not null;

alter table sessions
add column guest_data json null;

alter table sessions
add constraint sessions_kthid_or_guest_data check (
    (kthid is not null and guest_data is null) or
    (kthid is null and guest_data is not null)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table sessions
drop constraint sessions_kthid_or_guest_data;

alter table sessions
drop column guest_data;

alter table sessions
alter column kthid set not null;

alter table sessions
add constraint sessions_kthid_fkey foreign key (kthid) references users (kthid);
-- +goose StatementEnd
