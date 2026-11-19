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
create rule users_view_update as on update to users_view do instead (
    update users set
        email = new.email,
        first_name = new.first_name,
        family_name = new.family_name,
        year_tag = new.year_tag,
        webauthn_id = new.webauthn_id,
        first_name_change_request = new.first_name_change_request,
        family_name_change_request = new.family_name_change_request,
        nfc_id = new.nfc_id
    where kthid = old.kthid
    returning users.*, OLD.membership
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
