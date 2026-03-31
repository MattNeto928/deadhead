package search

import "github.com/mattneto928/deadhead/models"

// CitySummary holds the filtered flight results for a single destination city.
type CitySummary struct {
	Name              string
	FullName          string
	MinRoundTripPrice int
	MinLeavingPrice   int
	MinReturningPrice int
	Leaving           []*models.Flight
	Returning         []*models.Flight
}
