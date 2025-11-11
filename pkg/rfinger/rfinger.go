package rfinger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/datasektionen/sso/pkg/config"
)

func GetPicture(ctx context.Context, kthid string, quality bool) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(
		"%s/api/%s?quality=%t",
		config.Config.RfingerURL.String(),
		url.PathEscape(kthid),
		quality,
	), nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+config.Config.RfingerAPIKey)
	slog.Info("rfinger getPicture", "rfinger_url", config.Config.RfingerURL.String(), "url", req.URL)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unexpected status code from rfinger: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	body := string(bodyBytes)

	return body, nil
}

func GetPictures(ctx context.Context, kthid []string, quality bool) (map[string]string, error) {
	queryBody, err := json.Marshal(kthid)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(
		"%s/api/batch?quality=%t",
		config.Config.RfingerURL.String(),
		quality,
	), bytes.NewBuffer([]byte(queryBody)))

	req.Header.Set("Authorization", "Bearer "+config.Config.RfingerAPIKey)

	if err != nil {
		return nil, err
	}

	slog.Info("rfinger getPicture", "rfinger_url", config.Config.RfingerURL.String(), "url", req.URL)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code from rfinger: %d", resp.StatusCode)
	}

	var body map[string]string

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("Invalid JSON from rfinger: %w", err)
	}

	return body, nil
}
