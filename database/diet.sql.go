package database

import (
	"context"
	"database/sql"

	"github.com/datasektionen/sso/models"
	"github.com/google/uuid"
)

const listAllFoodPreferences = `-- name: ListAllFoodPreferences :many
select id, name_sv, name_en
from food_preference
`

func (q *Queries) ListAllFoodPreferences(ctx context.Context) ([]models.FoodPreference, error) {
	rows, err := q.db.Query(ctx, listAllFoodPreferences)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.FoodPreference
	for rows.Next() {
		var i models.FoodPreference
		if err := rows.Scan(
			&i.ID,
			&i.NameSv,
			&i.NameEn,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const userFoodPreferences = `-- name: UserFoodPreferences :many
select fp.id
from food_preference fp
join user_food_preference ufp on ufp.food_preference_id = fp.id
where ufp.kthid = $1
`

func (q *Queries) ListUserFoodPreferences(ctx context.Context, userID string) ([]uuid.UUID, error) {
	rows, err := q.db.Query(ctx, userFoodPreferences, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var i uuid.UUID
		if err := rows.Scan(&i); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllAllergies = `-- name: ListAllAllergies :many
select id, name_sv, name_en, parent_allergy_id, show_raw, show_severe, show_airborne
from allergy
`

func (q *Queries) ListAllAllergiesFlat(ctx context.Context) ([]models.Allergy, error) {
	rows, err := q.db.Query(ctx, listAllAllergies)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Allergy
	for rows.Next() {
		var i models.Allergy
		if err := rows.Scan(
			&i.ID,
			&i.NameSv,
			&i.NameEn,
			&i.ParentID,
			&i.ShowRaw,
			&i.ShowSevere,
			&i.ShowAirborne,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (q *Queries) ListAllAllergiesWithChildren(ctx context.Context) ([]models.Allergy, error) {
	allergies, err := q.ListAllAllergiesFlat(ctx)
	if err != nil {
		return nil, err
	}

	rootAllergies := make([]models.Allergy, 0)
	childrenOf := make(map[uuid.UUID][]models.Allergy)
	for _, allergy := range allergies {
		if allergy.ParentID == nil {
			rootAllergies = append(rootAllergies, allergy)
		} else {
			childrenOf[*allergy.ParentID] = append(childrenOf[*allergy.ParentID], allergy)
		}
	}

	var build func(allergy models.Allergy, depth int) models.Allergy
	build = func(allergy models.Allergy, depth int) models.Allergy {
		// Guard against cycles in the data.
		if depth > 32 {
			return allergy
		}
		for _, child := range childrenOf[allergy.ID] {
			allergy.Children = append(allergy.Children, build(child, depth+1))
		}
		return allergy
	}

	for i := range rootAllergies {
		rootAllergies[i] = build(rootAllergies[i], 0)
	}

	return rootAllergies, nil
}

const userAllergies = `-- name: UserAllergies :many
select a.id, ua.raw, ua.severe, ua.airborne
from allergy a
join user_allergy ua on ua.allergy_id = a.id
where ua.kthid = $1
`

func (q *Queries) ListUserAllergiesFlat(ctx context.Context, userID string) ([]models.UserAllergy, error) {
	rows, err := q.db.Query(ctx, userAllergies, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.UserAllergy
	for rows.Next() {
		var i models.UserAllergy
		if err := rows.Scan(&i.AllergyID, &i.Raw, &i.Severe, &i.Airborne); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const removeAllUserFoodPreferences = `-- name: RemoveAllUserFoodPreferences :exec
delete from user_food_preference
where kthid = $1
`

func (q *Queries) RemoveAllUserFoodPreferences(ctx context.Context, userID string) error {
	_, err := q.db.Exec(ctx, removeAllUserFoodPreferences, userID)
	return err
}

const addUserFoodPreferences = `-- name: AddUserFoodPreferences :exec
insert into user_food_preference (kthid, food_preference_id)
select $1, unnest($2::uuid[])
`

func (q *Queries) AddUserFoodPreferences(ctx context.Context, userID string, foodPreferenceIDs []uuid.UUID) error {
	_, err := q.db.Exec(ctx, addUserFoodPreferences, userID, foodPreferenceIDs)
	return err
}

const setUserFoodPreferenceFreetext = `-- name: SetUserFoodPreferenceFreetext :exec
update users set food_preference_freetext = $2 where kthid = $1;
`

func (q *Queries) SetUserFoodPreferenceFreetext(ctx context.Context, userID string, freetext string) error {
	_, err := q.db.Exec(ctx, setUserFoodPreferenceFreetext, userID, freetext)
	return err
}

const removeAllUserAllergies = `-- name: RemoveAllUserAllergies :exec
delete from user_allergy
where kthid = $1
`

func (q *Queries) RemoveAllUserAllergies(ctx context.Context, userID string) error {
	_, err := q.db.Exec(ctx, removeAllUserAllergies, userID)
	return err
}

const addUserAllergy = `-- name: AddUserAllergy :exec
insert into user_allergy (kthid, allergy_id, raw, severe, airborne)
values ($1, $2, $3, $4, $5)
`

func (q *Queries) AddUserAllergy(ctx context.Context, userID string, allergyID uuid.UUID, raw sql.NullBool, severe sql.NullBool, airborne sql.NullBool) error {
	_, err := q.db.Exec(ctx, addUserAllergy, userID, allergyID, raw, severe, airborne)
	return err
}

const setUserAllergyFreetext = `-- name: SetUserAllergyFreetext :exec
update users set allergy_freetext = $2 where kthid = $1;
`

func (q *Queries) SetUserAllergyFreetext(ctx context.Context, userID string, freetext string) error {
	_, err := q.db.Exec(ctx, setUserAllergyFreetext, userID, freetext)
	return err
}
