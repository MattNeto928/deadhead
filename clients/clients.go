package clients

import "net/http"

// Location represents a city with its associated airports.
type Location struct {
	City     string   `json:"city"`
	State    string   `json:"state"`
	Airports []string `json:"airports"`
}

// HTTPClient is the shared HTTP client used by all client functions.
// It is populated by Init() with Cloudflare session cookies and a matching
// User-Agent. It may be replaced in tests to point at a mock server.
var HTTPClient = &http.Client{}
