-- +goose Up
-- +goose StatementBegin
create table food_preference
(
    id      uuid default gen_random_uuid() not null
        primary key,
    name_sv varchar                        not null,
    name_en varchar                        not null
);

create table allergy
(
    id                uuid    default gen_random_uuid() not null
        primary key,
    name_sv           varchar                           not null,
    name_en           varchar                           not null,
    parent_allergy_id uuid
        references allergy,
    show_raw          boolean default false             not null,
    show_severe       boolean default false             not null,
    show_airborne     boolean default false             not null
);

create table user_allergy
(
    kthid      text not null
        references users
            on delete cascade,
    allergy_id uuid not null
        references allergy,
    raw        boolean,
    severe     boolean,
    airborne   boolean,
    primary key (kthid, allergy_id)
);

create table user_food_preference
(
    kthid              text not null
        references users
            on delete cascade,
    food_preference_id uuid not null
        references food_preference,
    primary key (kthid, food_preference_id)
);

alter table users add food_preference_freetext text not null default '';

alter table users add allergy_freetext text not null default '';

-- users_view expands users.* at creation time, so recreate it to include the new columns
drop view users_view;
create view users_view as (
    select users.*, active_membership.type as "membership"
    from users
    left join active_membership on users.kthid = active_membership.kthid
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table user_food_preference;
drop table user_allergy;
drop table allergy;
drop table food_preference;
drop view users_view;
alter table users drop column food_preference_freetext;
alter table users drop column allergy_freetext;
create view users_view as (
    select users.*, active_membership.type as "membership"
    from users
    left join active_membership on users.kthid = active_membership.kthid
);
-- +goose StatementEnd
