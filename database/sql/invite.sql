-- name: ListInvites :many
select * from invites;

-- name: GetInvite :one
select * from invites where id = $1;

-- name: CreateInvite :one
insert into invites (
    name,
    expires_at,
    max_uses
)
values ($1, $2, $3)
returning *;

-- name: DeleteInvite :exec
delete from invites
where id = $1;

-- name: IncrementInviteUses :exec
update invites
set current_uses = current_uses + 1
where id = $1;

-- name: UpdateInvite :one
update invites
set name = $2, expires_at = $3, max_uses = $4
where id = $1
returning *;
