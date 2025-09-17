package hive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/datasektionen/sso/pkg/config"
)

type PermissionScopes struct {
	Scopes []string
}

func (sp PermissionScopes) Matches(entity string) bool {
	for _, scope := range sp.Scopes {
		if base, ok := strings.CutSuffix(scope, "*"); ok {
			if strings.HasPrefix(entity, base) {
				return true
			}
		} else {
			if entity == scope {
				return true
			}
		}
	}
	return false
}

func (sp PermissionScopes) Exists() bool {
	return len(sp.Scopes) > 0
}

type Permissions struct {
	ReadMembers           bool             `hive:"read-members"`
	WriteMembers          bool             `hive:"write-members"`
	ReadOIDCClients       bool             `hive:"read-oidc-clients"`
	WriteOIDCClients      PermissionScopes `hive:"write-oidc-clients"`
	ReadInvites           bool             `hive:"read-invites"`
	WriteInvites          bool             `hive:"write-invites"`
	ReadAccountRequests   bool             `hive:"read-account-requests"`
	ManageAccountRequests bool             `hive:"manage-account-requests"`
}

type PermissionsCtxKey struct{}

func GetSSOPermissions(ctx context.Context, kthid string) (Permissions, error) {
	if config.Config.Dev && config.Config.HiveURL == nil {
		return Permissions{true, true, true, PermissionScopes{[]string{"*"}}, true, true, true, true}, nil
	}

	rawPerms, err := GetRawPermissionsInSystemForUser(ctx, kthid, "sso")
	if err != nil {
		return Permissions{}, err
	}

	var perms Permissions

	permType := reflect.TypeFor[Permissions]()
	permValue := reflect.ValueOf(&perms).Elem()
	for _, perm := range rawPerms {
		foundField := false
		for i := 0; i < permType.NumField(); i++ {
			field := permType.Field(i)
			tag := field.Tag.Get("hive")
			if tag != perm.ID {
				continue
			}
			foundField = true
			fieldValue := permValue.Field(i)

			if scopes, ok := fieldValue.Addr().Interface().(*PermissionScopes); ok {
				scopes.Scopes = append(scopes.Scopes, perm.Scope)
			} else if fieldValue.Type() == reflect.TypeFor[bool]() {
				fieldValue.SetBool(true)
			} else {
				panic("Unknown permission type")
			}
			break
		}
		if !foundField {
			return perms, fmt.Errorf("Got unknown permission from hive: '%s'", perm.ID)
		}
	}

	return perms, nil
}

type RawPermission struct {
	ID    string `json:"id"`
	Scope string `json:"scope"`
}

// Result follows the format described in the hive documentation. E.g.:
// [ { "id": "attest", "scope": "*" }, { "id": "view-logs", "scope": null }, { "id": "write", "scope": "/central/flag.txt" } ]
func GetRawPermissionsInSystemForUser(ctx context.Context, kthid string, system string) ([]RawPermission, error) {
	if config.Config.HiveURL == nil || config.Config.HiveAPIKey == "" {
		return nil, fmt.Errorf("URL or API key to hive missing in env variables, so cannot fetch permissions from hive")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(
		"%s/api/v1/user/%s/permissions",
		config.Config.HiveURL.String(),
		url.PathEscape(kthid),
	), nil)
	req.Header.Set("Authorization", "Bearer "+config.Config.HiveAPIKey)
	req.Header.Set("X-Hive-Impersonate-System", system)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code from hive: %d", resp.StatusCode)
	}
	var body []RawPermission
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("Invalid JSON from hive: %w", err)
	}
	return body, nil
}
