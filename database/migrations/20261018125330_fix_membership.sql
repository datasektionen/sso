-- +goose Up
-- +goose StatementBegin
drop view membership;

create view active_membership as (
    select kthid, max(type)::membership_type as "type"
    from memberships
    where end_date > current_date or end_date is null
    group by kthid
);

create view users_view as (
    select users.*, active_membership.type as "membership"
    from users
    left join active_membership on users.kthid = active_membership.kthid
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop view users_view;
drop view active_membership;

create view membership as (
    select kthid, max(type) as "type"
    from memberships
    where end_date > current_date
    group by kthid
);
-- +goose StatementEnd
