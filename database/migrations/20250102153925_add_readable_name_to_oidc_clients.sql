-- +goose Up
-- +goose StatementBegin
alter table oidc_clients rename column id to secret_hash;
alter table oidc_clients add column id text null;

-- Keep the ID's of the existing clients the same to not break anything. This is the same as `base64.URLEncoding`.
update oidc_clients
set id = replace(replace(encode(secret_hash, 'base64'), '+', '-'), '/', '_');

alter table oidc_clients alter column id set not null;

alter table oidc_clients drop constraint oidc_clients_pkey;
alter table oidc_clients add primary key (id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table oidc_clients drop column id;
alter table oidc_clients rename column secret_hash to id;
alter table oidc_clients add primary key (id);
-- +goose StatementEnd
