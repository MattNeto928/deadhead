package search

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/models"
)

// buildWorldSummaries converts a raw worldwide API response into a list of CitySummaries,
// applying the price filter from req. Separated for unit testing without HTTP.
func buildWorldSummaries(req *models.Request, manifest *clients.CountryResponse) []*CitySummary {
	byName := map[string]*CitySummary{}
	for _, trip := range manifest.Trips {
		price := trip.Cost / 100
		if req.Criteria.MaxPrice > 0 && price > req.Criteria.MaxPrice {
			continue
		}
		city := manifest.Cities[trip.City]
		fullName := fmt.Sprintf("%s, %s", city.Name, city.Region)
		if other, ok := byName[fullName]; !ok || other.MinRoundTripPrice > price {
			byName[fullName] = &CitySummary{
				Name:              trip.City,
				FullName:          fullName,
				MinRoundTripPrice: price,
			}
		}
	}

	summaries := make([]*CitySummary, 0, len(byName))
	for _, summary := range byName {
		summaries = append(summaries, summary)
	}
	return summaries
}

// GetCitySummaryLeavingCity fetches all possible destination cities and their minimum prices.
func GetCitySummaryLeavingCity(req *models.Request) ([]*CitySummary, error) {
	manifest, err := clients.GetWorldwideFlightsFromCity(req)
	if err != nil {
		return nil, errors.New("unable to get flights from city")
	}
	return buildWorldSummaries(req, manifest), nil
}

// StatusLogger is called by GetAllFlightSummariesToCity to report per-city progress.
// Set this to a function that writes to stderr (e.g. log.New(os.Stderr,"",0).Printf).
// If nil, progress is silenced.
var StatusLogger func(format string, args ...any)

// GetAllFlightSummariesToCity fetches detailed flight options for each city in the list,
// filters out cities with no qualifying flights, and returns results sorted by price.
func GetAllFlightSummariesToCity(req *models.Request, cities []*CitySummary) []*CitySummary {
	summaries := []*CitySummary{}
	total := len(cities)
	for i, city := range cities {
		req.TripCity = city.Name
		summary, err := GetFlightSummaryToCity(req)
		if err != nil {
			log.Println(err)
			if StatusLogger != nil {
				StatusLogger("[%d/%d] %s -- error: %v\n", i+1, total, city.Name, err)
			}
			continue
		}

		if len(summary.Leaving) == 0 {
			if StatusLogger != nil && total > 1 {
				StatusLogger("[%d/%d] %s -- no qualifying flights\n", i+1, total, city.Name)
			}
			continue
		}
		if !req.ReturningDay.IsZero() && len(summary.Returning) == 0 {
			if StatusLogger != nil && total > 1 {
				StatusLogger("[%d/%d] %s -- no qualifying return flights\n", i+1, total, city.Name)
			}
			continue
		}
		if StatusLogger != nil && total > 1 {
			minPrice := summary.MinLeavingPrice
			if !req.ReturningDay.IsZero() {
				minPrice = summary.MinRoundTripPrice
			}
			StatusLogger("[%d/%d] %s -- %d flight(s) from $%d\n", i+1, total, city.Name, len(summary.Leaving), minPrice)
		}
		summaries = append(summaries, summary)
		time.Sleep(time.Second * 2)
	}
	sort.Slice(summaries, func(i, j int) bool {
		if req.ReturningDay.IsZero() {
			return summaries[i].MinLeavingPrice < summaries[j].MinLeavingPrice
		}
		return summaries[i].MinRoundTripPrice < summaries[j].MinRoundTripPrice
	})
	return summaries
}
