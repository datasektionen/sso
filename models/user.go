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

type UserCtxKey struct{}

type GuestUserCtxKey struct{}
