package handlers

import (
	"net/http"
	"strconv"

	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
	"github.com/datasektionen/sso/models"
)

const memberSearchPageSize int32 = 12

func requireActiveMember(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if s.GetLoggedInGuestUser(r) != nil {
		return templates.MissingAccount()
	}
	user := s.GetLoggedInUser(r)
	if user == nil {
		s.RedirectToLogin(w, r, r.URL.Path)
		return nil
	}
	if !models.IsActiveMember(user) {
		return httputil.Forbidden("Only active members can search members")
	}
	return nil
}

func members(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if resp := requireActiveMember(s, w, r); resp != nil {
		return resp
	}
	years, err := s.DB.GetAllActiveMemberYears(r.Context())
	if err != nil {
		return err
	}
	return templates.MemberSearchPage(years)
}

func memberSearch(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	if resp := requireActiveMember(s, w, r); resp != nil {
		return resp
	}

	search := r.FormValue("search")
	year := r.FormValue("year")
	if len(search) < 2 && year == "" {
		return templates.MemberSearchResults(nil, map[string]string{}, search, year, 0, false)
	}

	offsetStr := r.FormValue("offset")
	var offset int64
	var err error
	if offsetStr != "" {
		offset, err = strconv.ParseInt(offsetStr, 10, 32)
		if err != nil {
			return httputil.BadRequest("Invalid int for offset")
		}
	}
	if offset < 0 {
		offset = 0
	}

	res, err := searchUsers(s, r, "search", memberSearchPageSize+1, int32(offset), true, true, false)
	if err != nil {
		return err
	}
	users := res.Users
	more := false
	if len(users) > int(memberSearchPageSize) {
		users = users[:memberSearchPageSize]
		more = true
	}
	return templates.MemberSearchResults(users, res.Pictures, search, year, int(offset), more)
}
