-- name: CreateToken :one
insert into legacyapi_tokens (kthid)
values ($1)
on conflict (kthid)
do update
set last_used_at = now()
returning id;

-- name: GetToken :one
update legacyapi_tokens
set last_used_at = now()
where id = $1 and last_used_at > now() - interval '8 hours'
returning kthid;

-- name: DeleteToken :exec
delete from legacyapi_tokens
where kthid = $1;
