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

-- name: CreateFoodPreference :exec
insert into food_preference (name_sv, name_en)
values ($1, $2);

-- name: UpdateFoodPreference :exec
update food_preference
set name_sv = $2, name_en = $3
where id = $1;

-- name: DeleteFoodPreference :exec
delete from food_preference
where id = $1;

-- name: CreateAllergy :exec
insert into allergy (parent_allergy_id, name_sv, name_en, show_raw, show_severe, show_airborne)
values ($1, $2, $3, $4, $5, $6);

-- name: UpdateAllergy :exec
update allergy
set parent_allergy_id = $2, name_sv = $3, name_en = $4, show_raw = $5, show_severe = $6, show_airborne = $7
where id = $1;

-- name: DeleteAllergy :exec
delete from allergy
where id = $1;