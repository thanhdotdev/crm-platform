package crmsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.com/bship1/bship-common-go.git/pkg/zutils_time"
)

// Client sends events to the CRM platform.
type Client struct {
	apiKey   string
	endpoint string
	http     *http.Client
}

// NewClient creates a new CRM SDK client.
// By default, reads CRM_API_KEY and CRM_ENDPOINT from environment variables.
func NewClient(opts ...Option) *Client {
	c := &Client{
		apiKey:   os.Getenv("CRM_API_KEY"),
		endpoint: os.Getenv("CRM_ENDPOINT"),
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.endpoint == "" {
		c.endpoint = "http://localhost:8080"
	}

	return c
}

// Track sends a single event to the CRM platform.
func (c *Client) Track(ctx context.Context, eventType, externalUserID, userType string, data map[string]interface{}) error {
	payload := map[string]interface{}{
		"event_type":       eventType,
		"external_user_id": externalUserID,
		"user_type":        userType,
		"data":             data,
		"timestamp":        zutils_time.NowGMT7().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("crmsdk: marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.endpoint+"/api/v1/ingest/events", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("crmsdk: request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("crmsdk: send error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("crmsdk: server returned %d", resp.StatusCode)
	}

	return nil
}
