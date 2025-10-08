package handlers

import (
	"net/http"
	"strconv"

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
		return httputil.JSON(convert(dbUsers[0]))
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
		return httputil.JSON(users)
	case "map":
		users := map[string]User{}
		for _, user := range dbUsers {
			users[user.Kthid] = convert(user)
		}

		return httputil.JSON(users)
	default:
		return httputil.BadRequest("Unknown or no data format requested")
	}
}

func apiSearchUsers(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	limitStr := r.FormValue("limit")
	if limitStr == "" {
		limitStr = "5"
	}
	i, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return err
	}
	limit := int32(i)

	offsetStr := r.FormValue("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	i, err = strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		return err
	}
	offset := int32(i)

	search := r.FormValue("query")
	year := r.FormValue("year")

	dbUsers, err := s.DB.ListUsers(r.Context(), database.ListUsersParams{
		Limit:  limit,
		Offset: offset,
		Search: search,
		Year:   year,
	})

	if err != nil {
		return err
	}

	type User struct {
		KTHID      string `json:"kthid"`
		Email      string `json:"email,omitempty"`
		FirstName  string `json:"firstName,omitempty"`
		FamilyName string `json:"familyName,omitempty"`
		YearTag    string `json:"yearTag,omitempty"`
	}

	users := make([]User, len(dbUsers))
	for i, user := range dbUsers {
		users[i] = User{
			KTHID:      user.Kthid,
			Email:      user.Email,
			FirstName:  user.FirstName,
			FamilyName: user.FamilyName,
			YearTag:    user.YearTag,
		}
	}
	return httputil.JSON(users)
}
