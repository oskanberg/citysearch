package cities_test

import (
	"reflect"
	"testing"

	"github.com/oskanberg/citysearch/cities"
)

func TestFilterCities(t *testing.T) {
	type test struct {
		name     string
		cities   []cities.City
		filters  []cities.FilterFunc
		expected []cities.City
	}

	cases := []test{
		{
			name: "doens't filter when none supplied",
			cities: []cities.City{
				{
					Name: "foo",
				},
			},
			expected: []cities.City{
				{
					Name: "foo",
				},
			},
		},
		{
			name: "removes non-GB when OnlyGB supplied",
			cities: []cities.City{
				{
					Name:        "foo",
					CountryCode: "GB",
				},
				{
					Name:        "bar",
					CountryCode: "DK",
				},
			},
			filters: []cities.FilterFunc{cities.OnlyGB},
			expected: []cities.City{
				{
					Name:        "foo",
					CountryCode: "GB",
				},
			},
		},
		{
			name: "runs multiple filters",
			cities: []cities.City{
				{
					Name:        "foo",
					CountryCode: "GB",
				},
				{
					Name:        "bar",
					CountryCode: "DK",
				},
				{
					Name:        "baz",
					CountryCode: "GB",
				},
			},
			filters: []cities.FilterFunc{
				cities.OnlyGB,
				func(c *cities.City) bool { return c.Name == "baz" },
			},
			expected: []cities.City{
				{
					Name:        "baz",
					CountryCode: "GB",
				},
			},
		},
	}

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := cities.Filter(tc.cities, tc.filters...)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
