package models

import "time"

type User struct {
	KTHID                   string
	UGKTHID                 string
	Email                   string
	FirstName               string
	FamilyName              string
	YearTag                 string
	MemberTo                time.Time
	WebAuthnID              []byte
	FirstNameChangeRequest  string
	FamilyNameChangeRequest string
}

type GuestUser struct {
	KTHID      string
	FirstName  string
	FamilyName string
}

func IsActiveMember(user *User) bool {
	if user == nil || user.MemberTo == (time.Time{}) {
		return false
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	memberTo := time.Date(user.MemberTo.Year(), user.MemberTo.Month(), user.MemberTo.Day(), 0, 0, 0, 0, now.Location())
	return !memberTo.Before(today)
}

type UserCtxKey struct{}

type GuestUserCtxKey struct{}
