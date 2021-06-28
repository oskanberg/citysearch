package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/oskanberg/citysearch/cities"
)

type CitySearcher interface {
	Search(ctx context.Context, query string) ([]cities.CityWithScore, error)
	SearchWithLocation(ctx context.Context, query string, lat, lng float64) ([]cities.CityWithScore, error)
}

type cityResult struct {
	Name string `json:"name"`

	// I noticed in the sample response that these are strings.
	// I think float64s make more sense - hope that's fair
	Lat float64 `json:"latitude"`
	Lng float64 `json:"longidude"`

	Score float64 `json:"score"`
}

type searchResultSDTO struct {
	Suggestions []cityResult `json:"suggestions"`
}

func NewCitySearchHandler(searcher CitySearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		params := r.URL.Query()
		query := params.Get("q")
		if query == "" {
			http.Error(w, "q (query string) must be set in URL", http.StatusBadRequest)
			return
		}

		lat, lng, locSet, err := getLatLng(params)
		if err != nil {
			http.Error(w, fmt.Sprintf("latitude/longitude error: %s", err.Error()), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		var result []cities.CityWithScore
		if locSet {
			result, err = searcher.SearchWithLocation(ctx, query, lat, lng)
		} else {
			result, err = searcher.Search(ctx, query)
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("search failed: %s", err), http.StatusInternalServerError)
			return
		}

		// construct dto
		cr := make([]cityResult, len(result))
		for i, v := range result {
			cr[i] = cityResult{
				Name:  v.Name,
				Lat:   v.Lat,
				Lng:   v.Lng,
				Score: v.Score,
			}
		}

		// make sure we are returning results highest score first
		sort.Slice(cr, func(i, j int) bool {
			return cr[i].Score > cr[j].Score
		})

		json.NewEncoder(w).Encode(searchResultSDTO{cr})
	}
}

func getLatLng(params url.Values) (float64, float64, bool, error) {
	latStr := params.Get("latitude")
	lngStr := params.Get("longitude")

	// if neither is set, no problem
	if latStr == "" && lngStr == "" {
		return 0, 0, false, nil
	}

	// at this point, one must be set; but check that they both are
	if latStr == "" || lngStr == "" {
		return 0, 0, false, fmt.Errorf("only one angle was provided")
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("latitude was not a number")
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return 0, 0, false, fmt.Errorf("longitude was not a number")
	}

	return lat, lng, true, nil
}
