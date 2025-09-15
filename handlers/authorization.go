package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/datasektionen/sso/pkg/hive"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
)

func authorize(s *service.Service, h http.Handler, permID string, scopeGetter func(*http.Request) string) http.Handler {
	return httputil.Route(s, func(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
		kthid, err := s.GetLoggedInKTHID(r)
		if err != nil {
			return err
		}
		if kthid == "" {
			s.RedirectToLogin(w, r, r.URL.Path)
			return nil
		}
		perms, err := hive.GetSSOPermissions(r.Context(), kthid)
		if err != nil {
			return err
		}
		permType := reflect.TypeFor[hive.Permissions]()
		permValue := reflect.ValueOf(&perms).Elem()
		var allowed, foundPerm bool
		for i := 0; i < permType.NumField(); i++ {
			field := permType.Field(i)
			tag := field.Tag.Get("hive")
			if tag != permID {
				continue
			}
			foundPerm = true
			fieldValue := permValue.Field(i)
			if scopes, ok := fieldValue.Addr().Interface().(*hive.PermissionScopes); ok {
				allowed = scopes.Matches(scopeGetter(r))
			} else if a, ok := fieldValue.Interface().(bool); ok {
				allowed = a
			} else {
				panic("Unknown permission type")
			}
			break
		}
		if !foundPerm {
			return fmt.Errorf("Tried checking for non-existing Hive permission with ID '%s'", permID)
		}
		if !allowed {
			return httputil.Forbidden("Missing Hive permission " + permID)
		}

		r = r.WithContext(context.WithValue(r.Context(), hive.Permissions{}, perms))

		h.ServeHTTP(w, r)

		if r.Method != http.MethodGet {
			args := []any{"kthid", kthid, "method", r.Method, "path", r.URL.Path}
			if id := r.PathValue("id"); id != "" {
				args = append(args, "id", id)
			} else if r.Form != nil && r.Form.Has("id") {
				args = append(args, "id", r.Form.Get("id"))
			}
			slog.InfoContext(r.Context(), "Admin action taken", args...)
		}

		return nil
	})
}
