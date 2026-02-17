package garminapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetProfile returns the authenticated user's profile settings.
func (c *Client) GetProfile(ctx context.Context) (json.RawMessage, error) {
	return c.ConnectAPI(ctx, http.MethodGet, "/userprofile-service/userprofile/settings", nil)
}

// GetUserSettings returns the authenticated user's settings.
func (c *Client) GetUserSettings(ctx context.Context) (json.RawMessage, error) {
	return c.ConnectAPI(ctx, http.MethodGet, "/userprofile-service/userprofile/user-settings", nil)
}

// GetDisplayName returns the authenticated user's display name from profile settings.
func (c *Client) GetDisplayName(ctx context.Context) (string, error) {
	data, err := c.ConnectAPI(ctx, http.MethodGet, "/userprofile-service/userprofile/settings", nil)
	if err != nil {
		return "", err
	}

	var settings struct {
		DisplayName string `json:"displayName"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return "", fmt.Errorf("parse profile settings: %w", err)
	}
	if settings.DisplayName == "" {
		return "", fmt.Errorf("profile settings missing displayName")
	}
	return settings.DisplayName, nil
}
