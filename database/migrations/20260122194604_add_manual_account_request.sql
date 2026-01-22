-- +goose Up
-- +goose StatementBegin
delete from account_requests
where kthid is null;

alter table account_requests
add column done boolean not null default true,
add column ug_kthid text not null default '',
add column first_name text not null default '',
add column family_name text not null default '',
add column email text not null default '',
alter column kthid set not null,
alter column kthid set default '';

alter table account_requests
alter column done drop default;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table account_requests
drop column ug_kthid,
drop column first_name,
drop column family_name,
drop column email,
drop column done,
alter column kthid drop not null,
alter column kthid drop default;
-- +goose StatementEnd
