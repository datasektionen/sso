package models

type User struct {
	KTHID                   string
	UGKTHID                 string
	Email                   string
	FirstName               string
	FamilyName              string
	YearTag                 string
	Membership              string
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
	return user != nil && user.Membership != "none"
}

type UserCtxKey struct{}

type GuestUserCtxKey struct{}
