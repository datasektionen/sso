package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
)

func apiListUsers(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	q := r.URL.Query()["u"]
	dbUsers, err := s.DB.GetUsersByIDs(r.Context(), q)
	if err != nil {
		return err
	}

	type User struct {
		Email      string `json:"email,omitempty"`
		FirstName  string `json:"firstName,omitempty"`
		FamilyName string `json:"familyName,omitempty"`
		YearTag    string `json:"yearTag,omitempty"`
	}
	convert := func(user database.User) User {
		return User{
			Email:      user.Email,
			FirstName:  user.FirstName,
			FamilyName: user.FamilyName,
			YearTag:    user.YearTag,
		}
	}

	switch r.URL.Query().Get("format") {
	case "single":
		if len(q) != 1 {
			return httputil.BadRequest("Single user requested but not exactly one username provided")
		}
		if len(dbUsers) != 1 {
			return httputil.NotFound()
		}
		return json.NewEncoder(w).Encode(convert(dbUsers[0]))
	case "array":
		indices := map[string]int{}
		for i, username := range q {
			if _, ok := indices[username]; ok {
				return httputil.BadRequest("Repeated username")
			}
			indices[username] = i
		}
		users := make([]User, len(q))
		for _, user := range dbUsers {
			users[indices[user.Kthid]] = convert(user)
		}
		return json.NewEncoder(w).Encode(users)
	case "map":
		users := map[string]User{}
		for _, user := range dbUsers {
			users[user.Kthid] = convert(user)
		}

		return json.NewEncoder(w).Encode(users)
	default:
		return httputil.BadRequest("Unknown or no data format requested")
	}
}
