-- name: GetClient :one
select *
from oidc_clients
where id = $1;

-- name: ListClients :many
select *
from oidc_clients;

-- name: CreateClient :one
insert into oidc_clients (id, secret_hash, redirect_uris)
values ($1, $2, '{}')
returning *;

-- name: UpdateClient :one
update oidc_clients
set redirect_uris = $2
where id = $1
returning *;

-- name: DeleteClient :exec
delete from oidc_clients
where id = $1;

-- name: AuthRequestSetKTHID :execrows
update oidcprovider_auth_requests
set kthid = $2
where id = $1;

-- name: GetAuthRequestByAuthCode :one
select *
from oidcprovider_auth_requests
where auth_code = $1;

-- name: GetAuthRequest :one
select *
from oidcprovider_auth_requests
where id = $1;

-- name: CreateAccessToken :one
insert into oidcprovider_access_tokens (kthid, scopes)
values ($1, $2)
returning id;

-- name: CreateAuthRequest :one
insert into oidcprovider_auth_requests (data, auth_code)
values ($1, '')
returning *;
