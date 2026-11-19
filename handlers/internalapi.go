package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/models"
	"github.com/datasektionen/sso/pkg/config"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/pkg/rfinger"
	"github.com/datasektionen/sso/service"
)

func apiListUsers(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	q := r.URL.Query()["u"]
	users, err := s.GetUsersByIDs(r.Context(), q)
	if err != nil {
		return err
	}

	nfcSearch := false

	for _, user := range users {
		if user.NFCID == q[0] {
			nfcSearch = true
			break
		}
	}

	type User struct {
		Kthid      string `json:"kthid,omitempty"`
		Email      string `json:"email,omitempty"`
		FirstName  string `json:"firstName,omitempty"`
		FamilyName string `json:"familyName,omitempty"`
		Picture    string `json:"picture,omitempty"`
		YearTag    string `json:"yearTag,omitempty"`
		Membership string `json:"membership,omitempty"`
	}
	getKTHID := func(users []models.User) []string {
		var kthids []string

		for _, user := range users {
			kthids = append(kthids, user.KTHID)
		}

		return kthids
	}
	convert := func(user models.User, picture string) User {
		return User{
			Kthid:      user.KTHID,
			Email:      user.Email,
			FirstName:  user.FirstName,
			FamilyName: user.FamilyName,
			Picture:    picture,
			YearTag:    user.YearTag,
			Membership: string(*user.Membership),
		}
	}

	pictures := make(map[string]string)

	if (r.FormValue("picture") == "full" || r.FormValue("picture") == "thumbnail") && len(users) > 0 {
		var err error
		if r.FormValue("format") == "single" {
			pictures[users[0].KTHID], err = rfinger.GetPicture(r.Context(), users[0].KTHID, r.FormValue("picture") == "full")
		} else {
			pictures, err = rfinger.GetPictures(r.Context(), getKTHID(users), r.FormValue("picture") == "full")
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
		if len(users) != 1 {
			return httputil.NotFound()
		}

		return httputil.JSON(convert(users[0], pictures[users[0].KTHID]))
	case "array":
		indices := map[string]int{}
		for i, username := range q {
			if _, ok := indices[username]; ok {
				return httputil.BadRequest("Repeated username")
			}
			indices[username] = i
		}
		convertedUsers := make([]User, len(q))

		if nfcSearch {
			for _, user := range users {
				convertedUsers[indices[user.NFCID]] = convert(user, pictures[user.KTHID])
			}
		} else {
			for _, user := range users {
				convertedUsers[indices[user.KTHID]] = convert(user, pictures[user.KTHID])
			}
		}
		return httputil.JSON(users)
	case "map":
		convertedUsers := map[string]User{}
		for _, user := range users {
			convertedUsers[user.KTHID] = convert(user, pictures[user.KTHID])
		}

		return httputil.JSON(users)
	default:
		return httputil.BadRequest("Unknown or no data format requested")
	}
}

type userSearchResult struct {
	Users    []models.User
	Pictures map[string]string
}

func searchUsers(s *service.Service, r *http.Request, searchParam string, limit int32, offset int32, membersOnly bool, includePictures bool, fullQualityPictures bool) (userSearchResult, error) {
	dbUsers, err := s.ListUsers(r.Context(), database.ListUsersParams{
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
			pictureUsers[i] = user.KTHID
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
		users[i] = User{
			KTHID:      user.KTHID,
			Email:      user.Email,
			FirstName:  user.FirstName,
			FamilyName: user.FamilyName,
			Picture:    res.Pictures[user.KTHID],
			YearTag:    user.YearTag,
			Membership: string(*user.Membership),
		}
	}
	return httputil.JSON(users)
}

type SetNFCIDParams struct {
	Kthid string `json:"kthid"`
	Nfcid string `json:"nfc_id"`
}

func apiSetNFCID(s *service.Service, _ http.ResponseWriter, r *http.Request) httputil.ToResponse {
	var params SetNFCIDParams
	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		return err
	}

	_, err = s.UserSetNFCID(r.Context(), params.Kthid, params.Nfcid)

	if err != nil {
		return err
	}

	return ""
}
