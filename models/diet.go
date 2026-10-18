package models

import (
	"database/sql"

	"github.com/google/uuid"
)

// TODO proper i18n stored somewhere else?
type Language string

const (
	LanguageSv Language = "sv"
	LanguageEn Language = "en"
)

type FoodPreference struct {
	ID     uuid.UUID `json:"id"`
	NameSv string    `json:"name_sv"`
	NameEn string    `json:"name_en"`
}

func (f FoodPreference) Name(language Language) string {
	if language == LanguageSv {
		return f.NameSv
	}
	return f.NameEn
}

type Allergy struct {
	ID           uuid.UUID  `json:"id"`
	NameSv       string     `json:"name_sv"`
	NameEn       string     `json:"name_en"`
	ShowRaw      bool       `json:"show_raw"`
	ShowSevere   bool       `json:"show_severe"`
	ShowAirborne bool       `json:"show_airborne"`
	ParentID     *uuid.UUID `json:"parent_id"`
	Children     []Allergy  `json:"children"`
}

func (a Allergy) Name(language Language) string {
	if language == LanguageSv {
		return a.NameSv
	}
	return a.NameEn
}

type UserAllergy struct {
	KTHID     string       `json:"kthid"`
	AllergyID uuid.UUID    `json:"id"`
	Raw       sql.NullBool `json:"raw"`
	Severe    sql.NullBool `json:"severe"`
	Airborne  sql.NullBool `json:"airborne"`
}
