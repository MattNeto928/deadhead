package search

import (
	"errors"
	"fmt"
	"time"

	"github.com/mattneto928/deadhead/clients"
	"github.com/mattneto928/deadhead/models"
)

// buildCitySummary processes a raw API response into a filtered CitySummary.
// It is separated from GetFlightSummaryToCity to allow unit testing without HTTP.
func buildCitySummary(req *models.Request, manifest *clients.CityResponse) *CitySummary {
	city := manifest.Info.To
	summary := CitySummary{
		Name:              req.TripCity,
		FullName:          fmt.Sprintf("%s, %s", city.City, city.State),
		MinRoundTripPrice: 0,
		MinLeavingPrice:   0,
		MinReturningPrice: 0,
		Leaving:           []*models.Flight{},
		Returning:         []*models.Flight{},
	}

	for _, outbound := range manifest.Itineraries.Outbound {
		flight, err := flightMeetsLeavingCriteria(manifest.Flights, outbound, req)
		if err != nil {
			continue
		}

		price := outbound.OneWayPrice / 100.0
		if summary.MinLeavingPrice == 0 || price < summary.MinLeavingPrice {
			summary.MinLeavingPrice = price
		}

		firstLeg := flight.Segments[0]
		lastLeg := flight.Segments[len(flight.Segments)-1]
		layovers := len(flight.Segments) - 1

		isHiddenCity := false
		hiddenDest := ""
		if layovers > 0 && lastLeg.Arrival.Airport != req.TripCity {
			// Basic heuristic: if it's a multi-leg flight and the final airport string doesn't
			// match our specific TripCity query string, it's likely a hidden-city ticket.
			isHiddenCity = true
			hiddenDest = lastLeg.Arrival.Airport
		}

		summary.Leaving = append(summary.Leaving, &models.Flight{
			Price:             price,
			Airline:           manifest.Airlines[firstLeg.Airline].Name,
			FlightNumber:      firstLeg.FlightNumber,
			Duration:          time.Duration(flight.Duration),
			Departure:         firstLeg.Departure,
			Arrival:           firstLeg.Arrival,
			IsHiddenCity:      isHiddenCity,
			HiddenDestination: hiddenDest,
			Layovers:          layovers,
		})
	}

	if len(summary.Leaving) > 0 && !req.ReturningDay.IsZero() {
		for _, inbound := range manifest.Itineraries.Inbound {
			flight, err := flightMeetsReturningCriteria(manifest.Flights, inbound, req)
			if err != nil {
				continue
			}

			price := inbound.OneWayPrice / 100.0
			if summary.MinReturningPrice == 0 || price < summary.MinReturningPrice {
				summary.MinReturningPrice = price
			}

			firstLeg := flight.Segments[0]
			lastLeg := flight.Segments[len(flight.Segments)-1]
			layovers := len(flight.Segments) - 1

			isHiddenCity := false
			hiddenDest := ""
			if layovers > 0 && lastLeg.Arrival.Airport != req.HomeCity {
				isHiddenCity = true
				hiddenDest = lastLeg.Arrival.Airport
			}

			summary.Returning = append(summary.Returning, &models.Flight{
				Price:             price,
				Airline:           manifest.Airlines[firstLeg.Airline].Name,
				FlightNumber:      firstLeg.FlightNumber,
				Duration:          time.Duration(flight.Duration),
				Departure:         firstLeg.Departure,
				Arrival:           firstLeg.Arrival,
				IsHiddenCity:      isHiddenCity,
				HiddenDestination: hiddenDest,
				Layovers:          layovers,
			})
		}
	}

	if len(summary.Leaving) > 0 && len(summary.Returning) > 0 {
		summary.MinRoundTripPrice = summary.MinLeavingPrice + summary.MinReturningPrice

		if req.Criteria.MaxPrice > 0 && summary.MinRoundTripPrice > req.Criteria.MaxPrice {
			summary.Leaving = []*models.Flight{}
			summary.Returning = []*models.Flight{}
		}
	}
	return &summary
}

// GetFlightSummaryToCity fetches and filters flight options to a specific city.
func GetFlightSummaryToCity(req *models.Request) (*CitySummary, error) {
	manifest, err := clients.GetFlightsToCity(req)
	if err != nil {
		return nil, errors.New("unable to get flights to city")
	}
	return buildCitySummary(req, manifest), nil
}
