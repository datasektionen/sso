package hive

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/datasektionen/sso/pkg/config"
)

type ScopedPermission struct {
	Scopes []string
}

func (sp ScopedPermission) Matches(entity string) bool {
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

type Permissions struct {
	ReadMembers           bool
	WriteMembers          bool
	ReadOIDCClients       bool
	WriteOIDCClients      ScopedPermission
	ReadInvites           bool
	WriteInvites          bool
	ReadAccountRequests   bool
	ManageAccountRequests bool
}

func GetSSOPermissions(ctx context.Context, kthid string) (Permissions, error) {
	rawPerms, err := GetRawPermissionsInSystemForUser(ctx, kthid, "sso")
	if err != nil {
		return Permissions{}, err
	}

	var perms Permissions
	for _, perm := range rawPerms {
		switch perm.ID {
		case "read-members":
			perms.ReadMembers = true
		case "write-members":
			perms.WriteMembers = true
		case "read-oidc-clients":
			perms.ReadOIDCClients = true
		case "write-oidc-clients":
			perms.WriteOIDCClients.Scopes = append(perms.WriteOIDCClients.Scopes, perm.Scope)
		case "read-invites":
			perms.ReadInvites = true
		case "write-invites":
			perms.WriteInvites = true
		case "read-account-requests":
			perms.ReadAccountRequests = true
		case "manage-account-requests":
			perms.ManageAccountRequests = true
		}
	}

	return perms, err
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
	slog.Info("hive getPermissions", "url", req)
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
