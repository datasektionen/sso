-- name: GetClient :one
select *
from oidc_clients
where id = $1;

-- name: ListClients :many
select *
from oidc_clients;

-- name: CreateClient :exec
insert into oidc_clients (id, redirect_uris)
values ($1, $2);

-- name: DeleteClient :exec
delete from oidc_clients
where id = $1;
