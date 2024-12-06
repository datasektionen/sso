package pls

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

func CheckUser(ctx context.Context, kthid, permission string) (bool, error) {
	return check(ctx, "user", kthid, permission)
}

func CheckToken(ctx context.Context, token, permission string) (bool, error) {
	return check(ctx, "token", token, permission)
}

func check(ctx context.Context, kind, who, permission string) (bool, error) {
	if config.Config.PlsURL == nil {
		return true, nil
	}

	system, perm, ok := strings.Cut(permission, ".")
	if !ok {
		system = "sso"
		perm = permission
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(
		"%s/api/%s/%s/%s/%s",
		config.Config.PlsURL.String(),
		url.PathEscape(kind),
		url.PathEscape(who),
		url.PathEscape(system),
		url.PathEscape(perm),
	), nil)
	if err != nil {
		return false, err
	}
	slog.Info("pls check", "pls_url", config.Config.PlsURL.String(), "url", req.URL)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		if kind == "token" {
			// Pls returns 500 when the token doesn't exist (what???)
			return false, nil
		} else {
			return false, fmt.Errorf("Unexpected status code from pls: %d", resp.StatusCode)
		}
	}
	var body bool
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("Invalid JSON from ldap proxy: %w", err)
	}
	return body, nil
}
