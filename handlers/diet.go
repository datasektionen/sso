package handlers

import (
	"database/sql"
	"net/http"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/models"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
	"github.com/google/uuid"
)

func updateFoodPreferences(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Invalid form body")
	}
	var ids []uuid.UUID
	for _, raw := range r.Form["food-preferences"] {
		id, err := uuid.Parse(raw)
		if err != nil {
			return httputil.BadRequest("Invalid food preference")
		}
		ids = append(ids, id)
	}
	freetext := r.FormValue("food-preference-freetext")
	if err := s.DB.Tx(r.Context(), func(db *database.Queries) error {
		if err := db.RemoveAllUserFoodPreferences(r.Context(), user.KTHID); err != nil {
			return err
		}
		if err := db.AddUserFoodPreferences(r.Context(), user.KTHID, ids); err != nil {
			return err
		}
		return db.SetUserFoodPreferenceFreetext(r.Context(), user.KTHID, freetext)
	}); err != nil {
		return err
	}
	allFoodPreferences, err := s.DB.ListAllFoodPreferences(r.Context())
	if err != nil {
		return err
	}
	return templates.FoodPreferences(models.LanguageSv, allFoodPreferences, ids, freetext, true)
}

func parseNullBool(s string) sql.NullBool {
	if s == "true" {
		return sql.NullBool{Bool: true, Valid: true}
	} else if s == "false" {
		return sql.NullBool{Bool: false, Valid: true}
	}
	return sql.NullBool{Valid: false}
}

func updateAllergies(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	user := s.GetLoggedInUser(r)
	if user == nil {
		return httputil.Unauthorized()
	}
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Invalid form body")
	}
	var userAllergies []models.UserAllergy
	for _, uuidStr := range r.Form["allergies"] {
		id, err := uuid.Parse(uuidStr)
		if err != nil {
			return httputil.BadRequest("Invalid allergy")
		}
		userAllergies = append(userAllergies, models.UserAllergy{
			KTHID:     user.KTHID,
			AllergyID: id,
			Raw:       parseNullBool(r.FormValue("allergy-" + uuidStr + "-raw")),
			Severe:    parseNullBool(r.FormValue("allergy-" + uuidStr + "-severe")),
			Airborne:  parseNullBool(r.FormValue("allergy-" + uuidStr + "-airborne")),
		})
	}
	freetext := r.FormValue("allergy-freetext")
	if err := s.DB.Tx(r.Context(), func(db *database.Queries) error {
		if err := db.RemoveAllUserAllergies(r.Context(), user.KTHID); err != nil {
			return err
		}
		for _, ua := range userAllergies {
			if err := db.AddUserAllergy(r.Context(), ua.KTHID, ua.AllergyID, ua.Raw, ua.Severe, ua.Airborne); err != nil {
				return err
			}
		}
		if err := db.SetUserAllergyFreetext(r.Context(), user.KTHID, freetext); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	allAllergies, err := s.DB.ListAllAllergiesWithChildren(r.Context())
	if err != nil {
		return err
	}
	userAllergiesFlat, err := s.DB.ListUserAllergiesFlat(r.Context(), user.KTHID)
	if err != nil {
		return err
	}
	return templates.Allergies(models.LanguageSv, allAllergies, userAllergiesFlat, freetext, true)
}
