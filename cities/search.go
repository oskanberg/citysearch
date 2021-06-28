package cities

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/jszwec/csvutil"
)

// stutters slightly, but naming is hard
type City struct {
	GeoNameID   string  `csv:"geonameid"`
	Name        string  `csv:"name"`
	Lat         float64 `csv:"latitude"`
	Lng         float64 `csv:"longitude"`
	CountryCode string  `csv:"country code"`
}

type CityWithScore struct {
	City
	Score float64
}

type CitySearcher struct {
	cities []City
}

// Filter is a type used to filter the database of cities
// Returning false implies the city should be discarded
type FilterFunc func(c *City) bool

// OnlyGB filters all locations that are not in GB country code
var OnlyGB FilterFunc = func(c *City) bool { return c.CountryCode == "GB" }

// Filter returns a (copy) slice with cities removed that did not pass
// all the provided filters
func Filter(cities []City, filters ...FilterFunc) []City {
	var filtered []City
	for _, c := range cities {
		var passed bool = true
		for _, f := range filters {
			// pointer to loop var usually dangerous, but ok here since it's just reading
			if !f(&c) {
				passed = false
				break
			}
		}
		if passed {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// NewCitySearcher creates a new CitySearcher with the given csv file reader
// and optional city filters
func NewCitySearcher(f io.Reader, filters ...FilterFunc) (*CitySearcher, error) {
	d, err := csvutil.NewDecoder(csv.NewReader(f))
	if err != nil {
		return nil, fmt.Errorf("failed to create csv decoder: %w", err)
	}

	var cities []City
	if err = d.Decode(&cities); err != nil {
		return nil, fmt.Errorf("failed to decode csv: %w", err)
	}

	return &CitySearcher{
		cities: Filter(cities, filters...),
	}, nil
}

// Search gets scored result suggestions for query
func (cs *CitySearcher) Search(ctx context.Context, query string) ([]CityWithScore, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// Search gets scored result suggestions for query, modulated by their proximity to lat/lng
func (cs *CitySearcher) SearchWithLocation(ctx context.Context, query string, lat, lng float64) ([]CityWithScore, error) {
	return nil, fmt.Errorf("not implemented yet")
}
