package cities

import (
	"context"
	"fmt"
)

// stutters slightly, but naming is hard
type City struct {
	GeoNameID   int
	Name        string
	Lat, Lng    float64
	CountryCode string
}

type CityWithScore struct {
	City
	Score float64
}

type CitySearcher struct{}

func NewCitySearcher() (*CitySearcher, error) {
	return &CitySearcher{}, nil
}

func (cs *CitySearcher) Search(ctx context.Context, query string) ([]CityWithScore, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (cs *CitySearcher) SearchWithLocation(ctx context.Context, query string, lat, lng float64) ([]CityWithScore, error) {
	return nil, fmt.Errorf("not implemented yet")
}
