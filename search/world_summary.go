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

// GetAllFlightSummariesToCity fetches detailed flight options for each city in the list,
// filters out cities with no qualifying flights, and returns results sorted by price.
func GetAllFlightSummariesToCity(req *models.Request, cities []*CitySummary) []*CitySummary {
	summaries := []*CitySummary{}
	for _, city := range cities {
		req.TripCity = city.Name
		summary, err := GetFlightSummaryToCity(req)
		if err != nil {
			log.Println(err)
			continue
		}

		if len(summary.Leaving) == 0 {
			continue
		}
		if !req.ReturningDay.IsZero() && len(summary.Returning) == 0 {
			continue
		}
		summaries = append(summaries, summary)
		time.Sleep(time.Second * 2)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].MinRoundTripPrice < summaries[j].MinRoundTripPrice
	})
	return summaries
}
