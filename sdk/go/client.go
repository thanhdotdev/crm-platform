package crmsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client is the CRM SDK client for sending events from backend services.
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

// Track sends an event to the CRM platform.
func (c *Client) Track(ctx context.Context, event Event) error {
	payload := eventPayload{
		EventType:      event.EventType(),
		ExternalUserID: event.UserID(),
		Data:           event,
		Timestamp:      time.Now(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("crmsdk: failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.endpoint+"/api/v1/ingest/events", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("crmsdk: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("crmsdk: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("crmsdk: server returned %d", resp.StatusCode)
	}

	return nil
}

// eventPayload is the JSON body sent to the Ingestion API.
type eventPayload struct {
	EventType      string    `json:"event_type"`
	ExternalUserID string    `json:"external_user_id"`
	Data           Event     `json:"data"`
	Timestamp      time.Time `json:"timestamp"`
}
