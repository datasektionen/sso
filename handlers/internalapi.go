package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/config"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/pkg/rfinger"
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
		Picture    string `json:"picture,omitempty"`
		YearTag    string `json:"yearTag,omitempty"`
		Membership string `json:"membership,omitempty"`
	}
	convert := func(user database.User, picture string) User {
		membership := "none"
		if user.Membership.Valid == true {
			membership = user.Membership.String
		}
		return User{
			Email:      user.Email,
			FirstName:  user.FirstName,
			FamilyName: user.FamilyName,
			Picture:    picture,
			YearTag:    user.YearTag,
			Membership: membership,
		}
	}

	pictures := make(map[string]string)

	if r.FormValue("picture") == "full" || r.FormValue("picture") == "thumbnail" {
		var err error
		if r.FormValue("format") == "single" {
			pictures[q[0]], err = rfinger.GetPicture(r.Context(), q[0], r.FormValue("picture") == "full")
		} else {
			pictures, err = rfinger.GetPictures(r.Context(), q, r.FormValue("picture") == "full")
		}

		if err != nil {
			return err
		}
	}

	switch r.FormValue("format") {
	case "single":
		if len(q) != 1 {
			return httputil.BadRequest("Single user requested but not exactly one username provided")
		}
		if len(dbUsers) != 1 {
			return httputil.NotFound()
		}

		return httputil.JSON(convert(dbUsers[0], pictures[dbUsers[0].Kthid]))
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
			users[indices[user.Kthid]] = convert(user, pictures[user.Kthid])
		}
		return httputil.JSON(users)
	case "map":
		users := map[string]User{}
		for _, user := range dbUsers {
			users[user.Kthid] = convert(user, pictures[user.Kthid])
		}

		return httputil.JSON(users)
	default:
		return httputil.BadRequest("Unknown or no data format requested")
	}
}

type userSearchResult struct {
	Users    []database.User
	Pictures map[string]string
}

func searchUsers(s *service.Service, r *http.Request, searchParam string, limit int32, offset int32, membersOnly bool, includePictures bool, fullQualityPictures bool) (userSearchResult, error) {
	dbUsers, err := s.DB.ListUsers(r.Context(), database.ListUsersParams{
		Limit:       limit,
		Offset:      offset,
		Search:      r.FormValue(searchParam),
		Year:        r.FormValue("year"),
		MembersOnly: membersOnly,
	})
	if err != nil {
		return userSearchResult{}, err
	}

	pictures := make(map[string]string, len(dbUsers))
	if includePictures && len(dbUsers) > 0 && config.Config.RfingerURL != nil {
		pictureUsers := make([]string, len(dbUsers))
		for i, user := range dbUsers {
			pictureUsers[i] = user.Kthid
		}
		pictures, err = rfinger.GetPictures(r.Context(), pictureUsers, fullQualityPictures)
		if err != nil {
			slog.WarnContext(r.Context(), "Could not fetch member pictures from rfinger", "error", err)
			pictures = map[string]string{}
		}
	}

	return userSearchResult{Users: dbUsers, Pictures: pictures}, nil
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

	res, err := searchUsers(s, r, "query", limit, offset, r.FormValue("membersOnly") == "true", r.FormValue("picture") == "full" || r.FormValue("picture") == "thumbnail", r.FormValue("picture") == "full")
	if err != nil {
		return err
	}

	type User struct {
		KTHID      string `json:"kthid"`
		Email      string `json:"email,omitempty"`
		FirstName  string `json:"firstName,omitempty"`
		FamilyName string `json:"familyName,omitempty"`
		Picture    string `json:"picture,omitempty"`
		YearTag    string `json:"yearTag,omitempty"`
		Membership string `json:"membership,omitempty"`
	}

	users := make([]User, len(res.Users))
	for i, user := range res.Users {
		membership := "none"
		if user.Membership.Valid == true {
			membership = user.Membership.String
		}
		users[i] = User{
			KTHID:      user.Kthid,
			Email:      user.Email,
			FirstName:  user.FirstName,
			FamilyName: user.FamilyName,
			Picture:    res.Pictures[user.Kthid],
			YearTag:    user.YearTag,
			Membership: membership,
		}
	}
	return httputil.JSON(users)
}
