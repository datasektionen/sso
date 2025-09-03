-- name: GetClient :one
select *
from oidc_clients
where id = $1;

-- name: ListClients :many
select *
from oidc_clients;

-- name: CreateClient :one
insert into oidc_clients (id, secret_hash, redirect_uris, hive_system_id)
values ($1, $2, '{}', $1)
returning *;

-- name: UpdateClientRedirectURIs :one
update oidc_clients
set redirect_uris = $2
where id = $1
returning *;

-- name: UpdateClientHiveSystemID :one
update oidc_clients
set hive_system_id = $2
where id = $1
returning *;

-- name: DeleteClient :exec
delete from oidc_clients
where id = $1;
