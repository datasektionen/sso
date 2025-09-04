package hive

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/datasektionen/sso/pkg/config"
)

// Result follows the format described in the hive documentation. E.g.:
// [ { "id": "attest", "scope": "*" }, { "id": "view-logs", "scope": null }, { "id": "write", "scope": "/central/flag.txt" } ]
func GetPermissionsInSystemForUser(ctx context.Context, kthid string, system string) (any, error) {
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
	var body any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("Invalid JSON from hive: %w", err)
	}
	return body, nil
}
