-- +goose Up
-- +goose StatementBegin
create type membership_type as enum ('honorary', 'junior', 'guest', 'alumni', 'ordinary');

create table memberships (
    kthid text references users (kthid),
    type membership_type not null,
    end_date date null,
    primary key (kthid, type),
    check (type = 'honorary' or type = 'alumni' or end_date is not null)
);

begin transaction;
    insert into memberships (kthid, end_date, type)
    select kthid, member_to, 'ordinary'
    from users
    where member_to is not null;
commit;

create view membership as (
    select kthid, max(type) as "type"
    from memberships
    where end_date > current_date
    group by kthid
);

alter table users drop column member_to;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users add column member_to date null;

begin transaction;
    update users
    set member_to = (
        select end_date
        from memberships
        where users.kthid = memberships.kthid
            and type = 'ordinary'
    );
commit;

drop table memberships;
drop type membership_type;
-- +goose StatementEnd
