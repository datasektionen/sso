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

-- name: GetUsersByIDs :many
select *
from users
where kthid = any(@ids::text[]);

-- name: ListUsers :many
select *
from users
where case
    when @search::text = '' then true
    else kthid = @search
      or first_name ~* @search
      or family_name ~* @search
      or first_name || ' ' || family_name ~* @search
end
and case
    when @year::text = '' then true
    else @year = year_tag
end
order by kthid
limit $1
offset $2;

-- name: GetAllYears :many
select distinct year_tag
from users
where year_tag != ''
order by year_tag;

-- name: UserSetMemberTo :exec
update users
set member_to = $2
where kthid = $1;

-- name: UserSetYear :one
update users
set year_tag = coalesce($2, year_tag)
where kthid = $1
returning *;

-- name: UserSetNameChangeRequest :one
update users
set first_name_change_request = $2,
    family_name_change_request = $3
where kthid = $1
returning *;

-- name: GetLastSheetUploadTime :one
select uploaded_at
from last_membership_sheet;

-- name: MarkSheetUploadedNow :exec
insert into last_membership_sheet (uploaded_at)
values (now())
on conflict (unique_marker)
do update
set uploaded_at = now();

-- name: CreateAccountRequest :one
insert into account_requests (reference, reason, year_tag)
values ($1, $2, $3)
returning id;

-- name: FinishAccountRequestKTH :exec
update account_requests
set kthid = $2
where id = $1;

-- name: ListAccountRequests :many
select *
from account_requests
order by created_at;

-- name: DeleteAccountRequest :one
delete from account_requests
where id = $1
returning *;
