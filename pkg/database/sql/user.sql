-- name: CreateSession :one
insert into sessions (kthid)
values ($1)
returning id;

-- name: GetSession :one
update sessions
set last_used_at = now()
where id = $1
and last_used_at > now() - interval '8 hours'
returning kthid;

-- name: RemoveSession :exec
delete from sessions
where id = $1;

-- name: CreateUser :exec
insert into users (kthid)
values ($1);

-- name: GetUser :one
select *
from users
where kthid = $1;
