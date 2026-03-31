package clients

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mattneto928/deadhead/models"
)

// CountryAPIBase is the base URL for the worldwide flight search endpoint.
// Override in tests to point at a mock server.
var CountryAPIBase = "https://skiplagged.com/api/skipsy.php"

// City represents a destination city returned by the worldwide search.
type City struct {
	Name     string   `json:"name"`
	Airports []string `json:"airports"`
	Region   string   `json:"region"`
}

// Trip represents a single destination option with pricing from the worldwide search.
type Trip struct {
	City        string `json:"city"`
	Cost        int    `json:"cost"`
	HiddenCity  bool   `json:"hidden_city"`
	RegularCost int    `json:"regular_cost,omitempty"`
}

// CountryResponse is the full API response for a worldwide flight search.
type CountryResponse struct {
	Cities   map[string]City `json:"cities"`
	Airports map[string]struct {
		Name string `json:"name"`
	} `json:"airports"`
	Info struct {
		From Location `json:"from"`
	} `json:"info"`
	Trips    []Trip  `json:"trips"`
	Duration float64 `json:"duration"`
}

// GetWorldwideFlightsFromCity fetches all possible destination cities from the origin.
func GetWorldwideFlightsFromCity(req *models.Request) (*CountryResponse, error) {
	returnDate := ""
	if !req.ReturningDay.IsZero() {
		returnDate = req.ReturningDay.Format("2006-01-02")
	}
	rawURL := fmt.Sprintf(
		"%s?from=%s&depart=%s&return=%s&format=v2&counts[adults]=%d&counts[children]=0&_=1611006103100",
		CountryAPIBase,
		url.QueryEscape(req.HomeCity),
		req.LeavingDay.Format("2006-01-02"),
		returnDate,
		req.Travelers,
	)
	res, err := HTTPClient.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var payload CountryResponse
	if err = json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
