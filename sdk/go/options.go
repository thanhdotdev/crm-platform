package crmsdk

// Option configures the CRM SDK client.
type Option func(*Client)

// WithAPIKey sets the API key directly (instead of reading from env).
func WithAPIKey(key string) Option {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithEndpoint sets the CRM API endpoint (instead of reading from env).
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}
