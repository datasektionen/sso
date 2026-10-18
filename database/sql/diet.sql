-- name: ListAllFoodPreferences :many
select *
from food_preference;

-- name: ListUserFoodPreferences :many
select food_preference_id
from user_food_preference
where kthid = $1;

-- name: ListAllAllergies :many
select *
from allergy;

-- name: ListUserAllergies :many
select allergy_id, raw, severe, airborne
from user_allergy
where kthid = $1;

-- name: RemoveAllUserFoodPreferences :exec
delete from user_food_preference
where kthid = $1;

-- name: AddUserFoodPreferences :exec
insert into user_food_preference (kthid, food_preference_id)
select @kthid, unnest(@food_preference_ids::uuid[]);

-- name: SetUserFoodPreferenceFreetext :exec
update users
set food_preference_freetext = $2
where kthid = $1;

-- name: RemoveAllUserAllergies :exec
delete from user_allergy
where kthid = $1;

-- name: AddUserAllergy :exec
insert into user_allergy (kthid, allergy_id, raw, severe, airborne)
values ($1, $2, $3, $4, $5);

-- name: SetUserAllergyFreetext :exec
update users
set allergy_freetext = $2
where kthid = $1;
