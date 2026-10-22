package handlers

import (
	"errors"
	"net/http"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func adminDietOptions(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	allFoodPreferences, err := s.ListAllFoodPreferences(r.Context())
	if err != nil {
		return err
	}
	allAllergiesFlat, err := s.ListAllAllergiesFlat(r.Context())
	if err != nil {
		return err
	}
	return templates.AdminDietOptions(allFoodPreferences, allAllergiesFlat)
}

func createFoodPreference(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Invalid form body")
	}
	nameSv := r.FormValue("name-sv")
	nameEn := r.FormValue("name-en")
	if nameSv == "" || nameEn == "" {
		return httputil.BadRequest("Missing name")
	}
	if err := s.DB.CreateFoodPreference(r.Context(), database.CreateFoodPreferenceParams{
		NameSv: nameSv,
		NameEn: nameEn,
	}); err != nil {
		return err
	}
	allFoodPreferences, err := s.ListAllFoodPreferences(r.Context())
	if err != nil {
		return err
	}
	return templates.AdminFoodPreferences(allFoodPreferences)
}

func updateFoodPreference(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Invalid form body")
	}
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httputil.BadRequest("Invalid food preference ID")
	}
	nameSv := r.FormValue("name-sv")
	nameEn := r.FormValue("name-en")
	if nameSv == "" || nameEn == "" {
		return httputil.BadRequest("Missing name")
	}
	if err := s.DB.UpdateFoodPreference(r.Context(), database.UpdateFoodPreferenceParams{
		ID:     id,
		NameSv: nameSv,
		NameEn: nameEn,
	}); err != nil {
		return err
	}
	allFoodPreferences, err := s.ListAllFoodPreferences(r.Context())
	if err != nil {
		return err
	}
	return templates.AdminFoodPreferences(allFoodPreferences)
}

func deleteFoodPreference(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httputil.BadRequest("Invalid food preference ID")
	}
	if err := s.DB.DeleteFoodPreference(r.Context(), id); err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23503" /* foreign_key_violation */ {
			return httputil.BadRequest("Cannot delete food preference since it is used")
		}
		return err
	}
	allFoodPreferences, err := s.ListAllFoodPreferences(r.Context())
	if err != nil {
		return err
	}
	return templates.AdminFoodPreferences(allFoodPreferences)
}

func createAllergy(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Invalid form body")
	}
	nameSv := r.FormValue("name-sv")
	nameEn := r.FormValue("name-en")
	parentIDStr := r.FormValue("parent-id")
	var parentID pgtype.UUID
	if parentIDStr != "null" && parentIDStr != "" {
		err := parentID.Scan(parentIDStr)
		if err != nil {
			return httputil.BadRequest("Invalid parent ID")
		}
	}
	if nameSv == "" || nameEn == "" {
		return httputil.BadRequest("Missing name")
	}
	if err := s.DB.CreateAllergy(r.Context(), database.CreateAllergyParams{
		ParentAllergyID: parentID,
		NameSv:          nameSv,
		NameEn:          nameEn,
		ShowRaw:         r.FormValue("show-raw") == "on",
		ShowSevere:      r.FormValue("show-severe") == "on",
		ShowAirborne:    r.FormValue("show-airborne") == "on",
	}); err != nil {
		return err
	}
	allAllergiesFlat, err := s.ListAllAllergiesFlat(r.Context())
	if err != nil {
		return err
	}
	return templates.AdminAllergies(allAllergiesFlat)
}

func updateAllergy(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if err := r.ParseForm(); err != nil {
		return httputil.BadRequest("Invalid form body")
	}
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httputil.BadRequest("Invalid allergy ID")
	}
	nameSv := r.FormValue("name-sv")
	nameEn := r.FormValue("name-en")
	parentIDStr := r.FormValue("parent-id")
	var parentID pgtype.UUID
	if parentIDStr != "null" && parentIDStr != "" {
		err := parentID.Scan(parentIDStr)
		if err != nil {
			return httputil.BadRequest("Invalid parent ID")
		}
	}
	if nameSv == "" || nameEn == "" {
		return httputil.BadRequest("Missing name")
	}
	if err := s.DB.Tx(r.Context(), func(db *database.Queries) error {
		if parentID.Valid {
			allergies, err := db.ListAllAllergies(r.Context())
			if err != nil {
				return err
			}
			if allergyParentCreatesCycle(allergies, id, parentID.Bytes) {
				return httputil.BadRequest("Cannot set parent since it would create a cycle")
			}
		}
		return db.UpdateAllergy(r.Context(), database.UpdateAllergyParams{
			ID:              id,
			NameSv:          nameSv,
			NameEn:          nameEn,
			ShowRaw:         r.FormValue("show-raw") == "on",
			ShowSevere:      r.FormValue("show-severe") == "on",
			ShowAirborne:    r.FormValue("show-airborne") == "on",
			ParentAllergyID: parentID,
		})
	}); err != nil {
		return err
	}
	allAllergiesFlat, err := s.ListAllAllergiesFlat(r.Context())
	if err != nil {
		return err
	}
	return templates.AdminAllergies(allAllergiesFlat)
}

func allergyParentCreatesCycle(allergies []database.Allergy, id uuid.UUID, parentID uuid.UUID) bool {
	parents := make(map[uuid.UUID]pgtype.UUID, len(allergies))
	for _, a := range allergies {
		parents[a.ID] = a.ParentAllergyID
	}
	visited := make(map[uuid.UUID]bool)
	for current := parentID; !visited[current]; {
		if current == id {
			return true
		}
		visited[current] = true
		parent := parents[current]
		if !parent.Valid {
			return false
		}
		current = parent.Bytes
	}
	return false
}

func deleteAllergy(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httputil.BadRequest("Invalid allergy ID")
	}
	if err := s.DB.DeleteAllergy(r.Context(), id); err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerr.Code == "23503" /* foreign_key_violation */ {
			return httputil.BadRequest("Cannot delete allergy since it is used")
		}
		return err
	}
	allAllergiesFlat, err := s.ListAllAllergiesFlat(r.Context())
	if err != nil {
		return err
	}
	return templates.AdminAllergies(allAllergiesFlat)
}
