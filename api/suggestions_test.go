package api_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oskanberg/citysearch/api"
	"github.com/oskanberg/citysearch/cities"
)

type mockSearcher struct {
	// records of arg calls
	query    string
	lat, lng float64

	// to return when invoked
	cities []cities.CityWithScore
}

func (cs *mockSearcher) Search(ctx context.Context, query string) ([]cities.CityWithScore, error) {
	cs.query = query
	return cs.cities, nil
}

func (cs *mockSearcher) SearchWithLocation(ctx context.Context, query string, lat, lng float64) ([]cities.CityWithScore, error) {
	cs.query = query
	cs.lat = lat
	cs.lng = lng
	return cs.cities, nil
}

func TestQueryParamsValidation(t *testing.T) {
	type test struct {
		name           string
		url            string
		expectedErr    string
		expectedStatus int
	}

	cases := []test{
		{
			name:           "no q set",
			url:            "/suggestions?",
			expectedErr:    "q (query string) must be set in URL\n",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "lat not number",
			url:            "/suggestions?q=foo&latitude=a&longitude=0.0",
			expectedErr:    "latitude/longitude error: latitude was not a number\n",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "lng not number",
			url:            "/suggestions?q=foo&latitude=0.0&longitude=a",
			expectedErr:    "latitude/longitude error: longitude was not a number\n",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "only lat set",
			url:            "/suggestions?q=foo&latitude=0.0",
			expectedErr:    "latitude/longitude error: only one angle was provided\n",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "only lng set",
			url:            "/suggestions?q=foo&longitude=0.0",
			expectedErr:    "latitude/longitude error: only one angle was provided\n",
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handle := api.NewCitySearchHandler(&mockSearcher{})
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, tc.url, nil)
			if err != nil {
				t.Fatalf("failure making test request: '%s'", err)
			}
			handle(rec, req)
			if rec.Code != tc.expectedStatus {
				t.Fatalf("expected response code '%d', got '%d'", tc.expectedStatus, rec.Code)
			}
			body, err := ioutil.ReadAll(rec.Body)
			if err != nil {
				t.Fatalf("failure getting test response: '%s'", err)
			}
			if string(body) != tc.expectedErr {
				t.Fatalf("expected response body '%s', got '%s'", tc.expectedErr, string(body))
			}
		})
	}
}

func TestCallsSearcherCorrectly(t *testing.T) {
	type test struct {
		name  string
		url   string
		check func(*testing.T, *mockSearcher)
	}

	cases := []test{
		{
			name: "calls without lat/lng",
			url:  "/suggestions?q=foo",
			check: func(t *testing.T, s *mockSearcher) {
				if s.query != "foo" {
					t.Fatalf("expected to call searcher with 'foo', but used '%s'", s.query)
				}
			},
		},
		{
			name: "calls with lat/lng",
			url:  "/suggestions?q=foo&latitude=0.1&longitude=0.2",
			check: func(t *testing.T, s *mockSearcher) {
				if s.query != "foo" {
					t.Fatalf("expected to call searcher with 'foo', but used '%s'", s.query)
				}
				if s.lat != 0.1 {
					t.Fatalf("expected to call searcher with lat 0.1, but used '%f'", s.lat)
				}
				if s.lng != 0.2 {
					t.Fatalf("expected to call searcher with lng 0.2, but used '%f'", s.lng)
				}
			},
		},
	}

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			searcher := &mockSearcher{}
			handle := api.NewCitySearchHandler(searcher)
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, tc.url, nil)
			if err != nil {
				t.Fatalf("failure making test request: '%s'", err)
			}
			handle(rec, req)
			if rec.Code != http.StatusOK {
				body, _ := ioutil.ReadAll(rec.Body)
				t.Fatalf("expected 200 OK but got %d ('%s')", rec.Code, string(body))
			}

			tc.check(t, searcher)
		})
	}
}

func TestResponseFormatting(t *testing.T) {
	type test struct {
		name           string
		searchResponse []cities.CityWithScore
		expectedBody   string
	}

	cases := []test{
		{
			name:           "empty response is empty array",
			searchResponse: []cities.CityWithScore{},
			expectedBody:   `{"suggestions":[]}`,
		},
		{
			name: "result includes expected properties",
			searchResponse: []cities.CityWithScore{{
				City: cities.City{
					GeoNameID: "1",
					Name:      "Wokingham",
					Lat:       51.4112,
					Lng:       -0.83565,
				},
				Score: 0.8,
			}},
			expectedBody: `{"suggestions":[{"name":"Wokingham","latitude":51.4112,"longitude":-0.83565,"score":0.8}]}`,
		},
		{
			name: "result sorted by score",
			searchResponse: []cities.CityWithScore{
				{
					City: cities.City{
						GeoNameID: "1",
						Name:      "Woking",
						Lat:       51.31903,
						Lng:       -0.55893,
					},
					Score: 0.6,
				},
				{
					City: cities.City{
						GeoNameID: "1",
						Name:      "Wokingham",
						Lat:       51.4112,
						Lng:       -0.83565,
					},
					Score: 0.8,
				},
			},
			expectedBody: `{"suggestions":[{"name":"Wokingham","latitude":51.4112,"longitude":-0.83565,"score":0.8},{"name":"Woking","latitude":51.31903,"longitude":-0.55893,"score":0.6}]}`,
		},
	}

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			searcher := &mockSearcher{
				cities: tc.searchResponse,
			}
			handle := api.NewCitySearchHandler(searcher)
			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/suggestions?q=a", nil)
			if err != nil {
				t.Fatalf("failure making test request: '%s'", err)
			}
			handle(rec, req)
			if rec.Code != http.StatusOK {
				body, _ := ioutil.ReadAll(rec.Body)
				t.Fatalf("expected 200 OK but got %d ('%s')", rec.Code, string(body))
			}

			b, err := ioutil.ReadAll(rec.Body)
			bStr := string(b)
			expected := tc.expectedBody + "\n"
			if bStr != expected {
				t.Fatalf("expected body '%s', but got '%s'", expected, bStr)
			}
		})
	}
}
