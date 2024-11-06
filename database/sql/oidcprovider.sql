-- name: GetClient :one
select *
from oidc_clients
where id = $1;

-- name: ListClients :many
select *
from oidc_clients;

-- name: CreateClient :one
insert into oidc_clients (id, redirect_uris)
values ($1, '{}')
returning *;

-- name: UpdateClient :one
update oidc_clients
set redirect_uris = $2
where id = $1
returning *;

-- name: DeleteClient :exec
delete from oidc_clients
where id = $1;
