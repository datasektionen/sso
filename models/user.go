package models

import "github.com/datasektionen/sso/database"

type User struct {
	KTHID                   string
	UGKTHID                 string
	Email                   string
	FirstName               string
	FamilyName              string
	YearTag                 string
	Membership              *database.MembershipType
	NFCID                   string
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
	return user != nil && user.Membership != nil
}

type UserCtxKey struct{}

type GuestUserCtxKey struct{}
