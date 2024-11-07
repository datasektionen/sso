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
insert into users (
    kthid,
    ug_kthid,
    email,
    first_name,
    family_name,
    year_tag,
    member_to
)
values ($1, $2, $3, $4, $5, $6, $7);

-- name: GetUser :one
select *
from users
where kthid = $1;

-- name: UserSetMemberTo :exec
update users
set member_to = $2
where kthid = $1;

-- name: UpdateUser :one
update users
set year_tag = coalesce($2, year_tag) -- TODO: would be nice if we could make this function take year_tag as a pointer so it can actually be null here
where kthid = $1
returning *;
