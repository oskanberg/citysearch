package cities

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/jszwec/csvutil"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/umahmood/haversine"
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
	cities    []City
	cityNames []string
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

	// filter the cities with filters provided
	cities = Filter(cities, filters...)
	if len(cities) == 0 {
		return nil, fmt.Errorf("no cities remained after filtering")
	}

	// keep a copy of the names, making sure they are in the same order
	cityNames := make([]string, len(cities))
	for i, c := range cities {
		cityNames[i] = strings.ToLower(c.Name)
	}

	return &CitySearcher{
		cities, cityNames,
	}, nil
}

// Search gets scored result suggestions for query. Scores are inversely proportional to the Levenshtein distance
func (cs *CitySearcher) Search(ctx context.Context, query string) ([]CityWithScore, error) {
	ranks := fuzzy.RankFind(strings.ToLower(query), cs.cityNames)
	result := make([]CityWithScore, len(ranks))
	for i, v := range ranks {
		result[i] = CityWithScore{
			// cs.cities is in the same order as cs.cityNames, so match by index
			City: cs.cities[v.OriginalIndex],
			// Levenshtein is 0 for a perfect match, so +1 to avoid /0
			Score: 1 / float64(v.Distance+1),
		}
	}

	return result, nil
}

// Search gets scored result suggestions for query, modulated by their proximity to lat/lng
func (cs *CitySearcher) SearchWithLocation(ctx context.Context, query string, lat, lng float64) ([]CityWithScore, error) {
	result, err := cs.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	distances := make([]float64, len(result))
	var min float64 = 1000 // 1000km is larger than GB
	for i, v := range result {
		// manhattan or euclidean distance wouldn't work here
		// since degrees in lat/lon are different sizes (and not all constant)
		_, km := haversine.Distance(
			haversine.Coord{Lat: lat, Lon: lng},
			haversine.Coord{Lat: v.Lat, Lon: v.Lng},
		)

		distances[i] = km
		if km < min {
			min = km
		}
	}

	for i, v := range result {
		// this is an arbitrary combining of the scores. future work
		// could improve this by tuning coefficiencs (or mixing nonlinearly)

		// 0 is best distance, add 1 to avoid /0
		distanceScore := math.Max(1, min) / (distances[i] + 1)
		result[i].Score = (v.Score + distanceScore) / 2
	}

	// re-sort taking into account distance scores
	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })

	return result, nil
}
