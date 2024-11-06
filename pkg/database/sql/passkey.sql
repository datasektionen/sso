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

-- name: StoreWebAuthnSessionData :exec
insert into webauthn_session_data (kthid, data)
values ($1, $2)
on conflict (kthid)
do update set data = $2;

-- name: TakeWebAuthnSessionData :one
delete from webauthn_session_data
where kthid = $1
returning data;
