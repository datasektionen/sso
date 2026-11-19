package service

import (
	"context"
	"database/sql"

	"github.com/datasektionen/sso/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) ListAllFoodPreferences(ctx context.Context) ([]models.FoodPreference, error) {
	dbFoodPreferences, err := s.DB.ListAllFoodPreferences(ctx)
	if err != nil {
		return nil, err
	}
	foodPreferences := make([]models.FoodPreference, len(dbFoodPreferences))
	for i, foodPreference := range dbFoodPreferences {
		foodPreferences[i] = models.FoodPreference{
			ID:     foodPreference.ID,
			NameSv: foodPreference.NameSv,
			NameEn: foodPreference.NameEn,
		}
	}
	return foodPreferences, nil
}

func (s *Service) ListAllAllergiesFlat(ctx context.Context) ([]models.Allergy, error) {
	dbAllergies, err := s.DB.ListAllAllergies(ctx)
	if err != nil {
		return nil, err
	}
	allergies := make([]models.Allergy, len(dbAllergies))
	for i, allergy := range dbAllergies {
		var parentID *uuid.UUID
		if allergy.ParentAllergyID.Valid {
			id := uuid.UUID(allergy.ParentAllergyID.Bytes)
			parentID = &id
		}
		allergies[i] = models.Allergy{
			ID:           allergy.ID,
			NameSv:       allergy.NameSv,
			NameEn:       allergy.NameEn,
			ShowRaw:      allergy.ShowRaw,
			ShowSevere:   allergy.ShowSevere,
			ShowAirborne: allergy.ShowAirborne,
			ParentID:     parentID,
		}
	}
	return allergies, nil
}

func (s *Service) ListAllAllergiesWithChildren(ctx context.Context) ([]models.Allergy, error) {
	allergies, err := s.ListAllAllergiesFlat(ctx)
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

func (s *Service) ListUserAllergiesFlat(ctx context.Context, kthid string) ([]models.UserAllergy, error) {
	dbUserAllergies, err := s.DB.ListUserAllergies(ctx, kthid)
	if err != nil {
		return nil, err
	}
	userAllergies := make([]models.UserAllergy, len(dbUserAllergies))
	for i, userAllergy := range dbUserAllergies {
		userAllergies[i] = models.UserAllergy{
			KTHID:     kthid,
			AllergyID: userAllergy.AllergyID,
			Raw:       pgBoolToNullBool(userAllergy.Raw),
			Severe:    pgBoolToNullBool(userAllergy.Severe),
			Airborne:  pgBoolToNullBool(userAllergy.Airborne),
		}
	}
	return userAllergies, nil
}

func pgBoolToNullBool(b pgtype.Bool) sql.NullBool {
	return sql.NullBool{Bool: b.Bool, Valid: b.Valid}
}

func NullBoolToPgBool(b sql.NullBool) pgtype.Bool {
	return pgtype.Bool{Bool: b.Bool, Valid: b.Valid}
}
