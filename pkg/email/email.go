package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/datasektionen/sso/pkg/config"
)

func Send(ctx context.Context, recipient, subject, contents string) error {
	slog.Info("Email sent", "recipient", recipient, "subject", subject)

	if config.Config.SpamURL == nil {
		slog.Info("Email not sent since no url was provided", "contents", contents)
		return nil
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(map[string]any{
		"key": config.Config.SpamAPIKey,
		"from": map[string]any{
			"name":    "Datasektionen SSO",
			"address": "no-reply@datasektionen.se",
		},
		"to":      []string{recipient},
		"subject": subject,
		"content": contents,
		"replyTo": "d-sys@datasektionen.se",
	}); err != nil {
		return err
	}
	resp, err := http.Post(config.Config.SpamURL.String()+"/api/legacy/sendmail", "application/json", &body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Could not send email with spam and could not get body: %w", err)
		}
		return fmt.Errorf("Could not send email with spam: %s", string(respBody))
	}
	return nil
}
