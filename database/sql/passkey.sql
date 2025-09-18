-- name: AddPasskey :one
insert into passkeys (kthid, name, data)
values ($1, $2, $3)
returning id;

-- name: RemovePasskey :exec
delete from passkeys
where kthid = $1 and id = $2;

-- name: ListPasskeysByUser :many
select *
from passkeys
where kthid = $1;

-- name: StoreWebAuthnSessionData :one
insert into webauthn_session_data (data, kthid)
values ($1, $2)
returning id;

-- name: TakeWebAuthnSessionData :one
delete from webauthn_session_data
where id = $1
returning data, kthid;
