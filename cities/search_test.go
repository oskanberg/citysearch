package cities_test

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
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

func TestParsingErrors(t *testing.T) {
	type test struct {
		name        string
		csv         string
		filter      cities.FilterFunc
		expectedErr error
	}

	cases := []test{
		{
			name:        "when empty",
			csv:         "",
			expectedErr: fmt.Errorf("failed to create csv decoder: EOF"),
		},
		{
			name:        "inconsistent rows",
			csv:         "a,b,c\n1",
			expectedErr: fmt.Errorf("failed to decode csv: record on line 2: wrong number of fields"),
		},
		{
			name:        "no cities remain after filtering",
			csv:         "geonameid,name\n1,foo\n2,bar",
			filter:      func(c *cities.City) bool { return false },
			expectedErr: fmt.Errorf("no cities remained after filtering"),
		},
	}

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := cities.NewCitySearcher(strings.NewReader(tc.csv), tc.filter)
			// Sprintf'ing means it can also check nils
			if fmt.Sprintf("%s", err) != fmt.Sprintf("%s", tc.expectedErr) {
				t.Fatalf("expected error '%s', but got '%s'", fmt.Sprintf("%s", tc.expectedErr), fmt.Sprintf("%s", err))
			}
		})
	}
}

func TestSearch(t *testing.T) {
	citySample := `geonameid,name,asciiname,alternatenames,latitude,longitude,feature class,feature code,country code,cc2,admin1 code,admin2 code,admin3 code,admin4 code,population,elevation,dem,timezone,modification date
2633485,Wrexham,Wrexham,"Reksamas,Reksem,Reksum,Rexam,Wrecsam,Wreksam,Wrexham,legseom,lei ke si han mu,rekusamu,wrksam,Œ°Œ≠ŒæŒ±Œº,–†–µ–∫—Å–µ–º,–†–µ–∫—Å—ä–º,’å’•÷Ñ’Ω’∞’•’¥,◊®◊ß◊°◊î◊ê◊ù,Ÿàÿ±⁄©ÿ≥ÿßŸÖ,„É¨„ÇØ„Çµ„É†,Èõ∑ÂÖãÊñØÊº¢ÂßÜ,Î†âÏÑ¨",53.04664,-2.99132,P,PPLA2,GB,,WLS,Z4,00NL007,,65692,,87,Europe/London,12/06/2017
2633521,Worthing,Worthing,"Vorting,Worthing,wajingu,wwrtyng,–í–æ—Ä—Ç–∏–Ω–≥,ŸàŸàÿ±ÿ™€åŸÜ⁄Ø,„ÉØ„Éº„Ç∏„É≥„Ç∞",50.81795,-0.37538,P,PPL,GB,,ENG,P6,45UH,,99110,,7,Europe/London,14/09/2014
2633551,Worksop,Worksop,"Uehrksop,Uurksop,Worksop,wwrksap,–£—ä—Ä–∫—Å–æ–ø,–£—ç—Ä–∫—Å–æ–ø,ŸàŸàÿ±⁄©ÿ≥ÿßŸæ",53.30182,-1.12404,P,PPL,GB,,ENG,J9,37UC,,43252,,46,Europe/London,12/06/2017
2633553,Workington,Workington,"Uurkingtun,wo jin dun,wwrkyngtwn,–£—ä—Ä–∫–∏–Ω–≥—Ç—ä–Ω,ŸàŸàÿ±⁄©€åŸÜ⁄Øÿ™ŸàŸÜ,Ê≤ÉÈáëÈ†ì",54.6425,-3.54413,P,PPL,GB,,ENG,C9,16UB,16UB061,27120,,20,Europe/London,03/07/2018
2633563,Worcester,Worcester,"Caerwrangon,City of Worcester,UWC,Ustur,Vigornia,Vustehr,Vuster,Vusteris,Wiogoraceastre,Worcester,useuteo,usuta,vuster,wo shi da,wrkstr,wstr,wu si te,wurs texr,wwstr,–í—É—Å—Ç–µ—Ä,–í—É—Å—Ç—ç—Ä,–£—Å—Ç—ä—Ä,’é’∏÷Ç’Ω’©’•÷Ä,Ÿàÿ±⁄©ÿ≥ÿ™ÿ±,Ÿàÿ≥ÿ™ÿ±,ŸàŸàÿ±ÿ≥ÿ≥Ÿπÿ±,ŸàŸàÿ≥ÿ™ÿ±,ŸàŸàÿ≥Ÿπÿ±,‡∏ß‡∏∏‡∏£‡πå‡∏™‡πÄ‡∏ï‡∏≠‡∏£‡πå,„Ç¶„Çπ„Çø„Éº,‰ºçÊñØÁâπ,Á™©Â£´Êâì,Ïö∞Ïä§ÌÑ∞",52.18935,-2.22001,P,PPLA2,GB,,ENG,Q4,47UE,,101659,,29,Europe/London,05/09/2019
2633655,Woodford Green,Woodford Green,"Woodford Green,vudafarda grina,‡§µ‡•Å‡§°‡§´‡§º‡§∞‡•ç‡§° ‡§ó‡•ç‡§∞‡•Ä‡§®",51.60938,0.02329,P,PPL,GB,,ENG,GLA,K8,,22803,,62,Europe/London,18/05/2012
2633681,Wombwell,Wombwell,Wombwell,53.52189,-1.39698,P,PPL,GB,,ENG,A3,,,15518,,53,Europe/London,11/07/2013
2633708,Wokingham,Wokingham,"Uokingam,Uokingkhem,oking-eom,wwkyngham,–£–æ–∫–∏–Ω–≥–∞–º,–£–æ–∫–∏–Ω–≥—Ö–µ–º,ŸàŸà⁄©€åŸÜ⁄ØŸáÿßŸÖ,Ïò§ÌÇπÏóÑ",51.4112,-0.83565,P,PPLA2,GB,,ENG,Q2,00MF015,,41143,,72,Europe/London,22/06/2016
2633709,Woking,Woking,"Uoking,Uokinge,Vokingas,Woking,XWO,u~okingu,wo jin,wwdkyng,wwkng,–£–æ–∫–∏–Ω–≥,–£–æ–∫–∏–Ω–≥–µ,ŸàŸàÿØ⁄©€åŸÜ⁄Ø,ŸàŸà⁄©ŸÜ⁄Ø,„Ç¶„Ç©„Ç≠„É≥„Ç∞,Ê≤ÉÈáë",51.31903,-0.55893,P,PPL,GB,,ENG,N7,43UM,,103932,,39,Europe/London,03/08/2010
2633729,Witney,Witney,"Uitni,Witney,wytny,–£–∏—Ç–Ω–∏,Ÿà€åÿ™ŸÜ€å",51.7836,-1.4854,P,PPL,GB,,ENG,K2,38UF,38UF080,29103,,87,Europe/London,03/07/2018
2633749,Witham,Witham,"wytham,Ÿà€åÿ™ŸáÿßŸÖ",51.80007,0.64038,P,PPL,GB,,ENG,E4,22UC,22UC063,25353,,25,Europe/London,03/07/2018
2633765,Wishaw,Wishaw,"Camas Neachdain,Vishou,Wishae,Wishaw,wei xiao,wyshaw,–í—ñ—à–æ—É,Ÿà€åÿ¥ÿßŸà,Â®ÅËï≠",55.76667,-3.91667,P,PPL,GB,,SCT,V8,,,30510,,138,Europe/London,12/06/2017
2633771,Wisbech,Wisbech,"Uisbijch,Visbicas,Visbiƒças,Vizbich,Wisbech,wysbch,–í–∏–∑–±–∏—á,–£–∏—Å–±–∏–π—á,Ÿà€åÿ≥ÿ®⁄Ü",52.66622,0.15938,P,PPL,GB,,ENG,C3,12UD,12UD014,32489,,6,Europe/London,03/07/2018
2633810,Winsford,Winsford,"wynsfwrd,Ÿà€åŸÜÿ≥ŸÅŸàÿ±ÿØ",53.19146,-2.52398,P,PPLA3,GB,,ENG,Z8,00EW163,,30259,,36,Europe/London,13/06/2017
`

	type test struct {
		name  string
		csv   string
		query string
		check func(t *testing.T, c []cities.CityWithScore)
	}

	cases := []test{
		{
			name:  "exact match, score 1",
			csv:   citySample,
			query: "Wrexham",
			check: func(t *testing.T, c []cities.CityWithScore) {
				if c[0].Name != "Wrexham" {
					t.Fatalf("expected the first result to be Wrexham, but was %s", c[0].Name)
				}
				if c[0].Score != 1.0 {
					t.Fatalf("expected the first result to have score 1.0, but got %f", c[0].Score)
				}
			},
		},
		{
			name:  "also score 1 regardless of case",
			csv:   citySample,
			query: "wrexHAM",
			check: func(t *testing.T, c []cities.CityWithScore) {
				if c[0].Name != "Wrexham" {
					t.Fatalf("expected the first result to be Wrexham, but was %s", c[0].Name)
				}
				if c[0].Score != 1.0 {
					t.Fatalf("expected the first result to have score 1.0, but got %f", c[0].Score)
				}
			},
		},
		{
			name:  "worse match has lower score",
			csv:   citySample,
			query: "woon", // a few names have these letters
			check: func(t *testing.T, c []cities.CityWithScore) {
				if c[0].Name != "Workington" {
					t.Fatalf("expected the first result to be Workington, but was %s", c[0].Name)
				}

				// Woodford green is a worse match because it is longer
				if c[1].Name != "Woodford Green" {
					t.Fatalf("expected the second result to be Woodford Green, but was %s", c[1].Name)
				}

				// the scores themselves are kind of arbitrary,
				// so just check that they are ranked correctly
				if c[1].Score >= c[0].Score {
					t.Fatal("expected better match to have higher score")
				}
			},
		},
	}

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cs, err := cities.NewCitySearcher(strings.NewReader(tc.csv))
			if err != nil {
				t.Fatalf("failed to make city searcher: %s", err)
			}

			results, err := cs.Search(context.Background(), tc.query)
			if err != nil {
				t.Fatalf("failed to  search: %s", err)
			}

			tc.check(t, results)
		})
	}
}

func TestSearchWithLocation(t *testing.T) {
	citySample := `geonameid,name,asciiname,alternatenames,latitude,longitude,feature class,feature code,country code,cc2,admin1 code,admin2 code,admin3 code,admin4 code,population,elevation,dem,timezone,modification date
2633485,Wrexham,Wrexham,"Reksamas,Reksem,Reksum,Rexam,Wrecsam,Wreksam,Wrexham,legseom,lei ke si han mu,rekusamu,wrksam,Œ°Œ≠ŒæŒ±Œº,–†–µ–∫—Å–µ–º,–†–µ–∫—Å—ä–º,’å’•÷Ñ’Ω’∞’•’¥,◊®◊ß◊°◊î◊ê◊ù,Ÿàÿ±⁄©ÿ≥ÿßŸÖ,„É¨„ÇØ„Çµ„É†,Èõ∑ÂÖãÊñØÊº¢ÂßÜ,Î†âÏÑ¨",53.04664,-2.99132,P,PPLA2,GB,,WLS,Z4,00NL007,,65692,,87,Europe/London,12/06/2017
2633521,Worthing,Worthing,"Vorting,Worthing,wajingu,wwrtyng,–í–æ—Ä—Ç–∏–Ω–≥,ŸàŸàÿ±ÿ™€åŸÜ⁄Ø,„ÉØ„Éº„Ç∏„É≥„Ç∞",50.81795,-0.37538,P,PPL,GB,,ENG,P6,45UH,,99110,,7,Europe/London,14/09/2014
2633551,Worksop,Worksop,"Uehrksop,Uurksop,Worksop,wwrksap,–£—ä—Ä–∫—Å–æ–ø,–£—ç—Ä–∫—Å–æ–ø,ŸàŸàÿ±⁄©ÿ≥ÿßŸæ",53.30182,-1.12404,P,PPL,GB,,ENG,J9,37UC,,43252,,46,Europe/London,12/06/2017
2633553,Workington,Workington,"Uurkingtun,wo jin dun,wwrkyngtwn,–£—ä—Ä–∫–∏–Ω–≥—Ç—ä–Ω,ŸàŸàÿ±⁄©€åŸÜ⁄Øÿ™ŸàŸÜ,Ê≤ÉÈáëÈ†ì",54.6425,-3.54413,P,PPL,GB,,ENG,C9,16UB,16UB061,27120,,20,Europe/London,03/07/2018
2633563,Worcester,Worcester,"Caerwrangon,City of Worcester,UWC,Ustur,Vigornia,Vustehr,Vuster,Vusteris,Wiogoraceastre,Worcester,useuteo,usuta,vuster,wo shi da,wrkstr,wstr,wu si te,wurs texr,wwstr,–í—É—Å—Ç–µ—Ä,–í—É—Å—Ç—ç—Ä,–£—Å—Ç—ä—Ä,’é’∏÷Ç’Ω’©’•÷Ä,Ÿàÿ±⁄©ÿ≥ÿ™ÿ±,Ÿàÿ≥ÿ™ÿ±,ŸàŸàÿ±ÿ≥ÿ≥Ÿπÿ±,ŸàŸàÿ≥ÿ™ÿ±,ŸàŸàÿ≥Ÿπÿ±,‡∏ß‡∏∏‡∏£‡πå‡∏™‡πÄ‡∏ï‡∏≠‡∏£‡πå,„Ç¶„Çπ„Çø„Éº,‰ºçÊñØÁâπ,Á™©Â£´Êâì,Ïö∞Ïä§ÌÑ∞",52.18935,-2.22001,P,PPLA2,GB,,ENG,Q4,47UE,,101659,,29,Europe/London,05/09/2019
2633655,Woodford Green,Woodford Green,"Woodford Green,vudafarda grina,‡§µ‡•Å‡§°‡§´‡§º‡§∞‡•ç‡§° ‡§ó‡•ç‡§∞‡•Ä‡§®",51.60938,0.02329,P,PPL,GB,,ENG,GLA,K8,,22803,,62,Europe/London,18/05/2012
2633681,Wombwell,Wombwell,Wombwell,53.52189,-1.39698,P,PPL,GB,,ENG,A3,,,15518,,53,Europe/London,11/07/2013
2633708,Wokingham,Wokingham,"Uokingam,Uokingkhem,oking-eom,wwkyngham,–£–æ–∫–∏–Ω–≥–∞–º,–£–æ–∫–∏–Ω–≥—Ö–µ–º,ŸàŸà⁄©€åŸÜ⁄ØŸáÿßŸÖ,Ïò§ÌÇπÏóÑ",51.4112,-0.83565,P,PPLA2,GB,,ENG,Q2,00MF015,,41143,,72,Europe/London,22/06/2016
2633709,Woking,Woking,"Uoking,Uokinge,Vokingas,Woking,XWO,u~okingu,wo jin,wwdkyng,wwkng,–£–æ–∫–∏–Ω–≥,–£–æ–∫–∏–Ω–≥–µ,ŸàŸàÿØ⁄©€åŸÜ⁄Ø,ŸàŸà⁄©ŸÜ⁄Ø,„Ç¶„Ç©„Ç≠„É≥„Ç∞,Ê≤ÉÈáë",51.31903,-0.55893,P,PPL,GB,,ENG,N7,43UM,,103932,,39,Europe/London,03/08/2010
2633729,Witney,Witney,"Uitni,Witney,wytny,–£–∏—Ç–Ω–∏,Ÿà€åÿ™ŸÜ€å",51.7836,-1.4854,P,PPL,GB,,ENG,K2,38UF,38UF080,29103,,87,Europe/London,03/07/2018
2633749,Witham,Witham,"wytham,Ÿà€åÿ™ŸáÿßŸÖ",51.80007,0.64038,P,PPL,GB,,ENG,E4,22UC,22UC063,25353,,25,Europe/London,03/07/2018
2633765,Wishaw,Wishaw,"Camas Neachdain,Vishou,Wishae,Wishaw,wei xiao,wyshaw,–í—ñ—à–æ—É,Ÿà€åÿ¥ÿßŸà,Â®ÅËï≠",55.76667,-3.91667,P,PPL,GB,,SCT,V8,,,30510,,138,Europe/London,12/06/2017
2633771,Wisbech,Wisbech,"Uisbijch,Visbicas,Visbiƒças,Vizbich,Wisbech,wysbch,–í–∏–∑–±–∏—á,–£–∏—Å–±–∏–π—á,Ÿà€åÿ≥ÿ®⁄Ü",52.66622,0.15938,P,PPL,GB,,ENG,C3,12UD,12UD014,32489,,6,Europe/London,03/07/2018
2633810,Winsford,Winsford,"wynsfwrd,Ÿà€åŸÜÿ≥ŸÅŸàÿ±ÿØ",53.19146,-2.52398,P,PPLA3,GB,,ENG,Z8,00EW163,,30259,,36,Europe/London,13/06/2017
`

	type test struct {
		name     string
		csv      string
		query    string
		lat, lng float64
		check    func(t *testing.T, c []cities.CityWithScore)
	}

	cases := []test{
		{
			name:  "exact match, score 1",
			csv:   citySample,
			query: "Wrexham",
			lat:   53.04664,
			lng:   -2.99132,
			check: func(t *testing.T, c []cities.CityWithScore) {
				if c[0].Name != "Wrexham" {
					t.Fatalf("expected the first result to be Wrexham, but was %s", c[0].Name)
				}
				if c[0].Score != 1.0 {
					t.Fatalf("expected the first result to have score 1.0, but got %f", c[0].Score)
				}
			},
		},
		{
			name:  "exact string match, but very far away: score ~0.5",
			csv:   citySample,
			query: "Wrexham",
			lat:   0,
			lng:   0,
			check: func(t *testing.T, c []cities.CityWithScore) {
				if c[0].Name != "Wrexham" {
					t.Fatalf("expected the first result to be Wrexham, but was %s", c[0].Name)
				}
				if math.Abs(0.5-c[0].Score) > 1e-1 {
					t.Fatalf("expected the first result to have score 0.5, but got %f", c[0].Score)
				}
			},
		},
		{
			name:  "similar scoring names are ranked by distance",
			csv:   citySample,
			query: "i", // there are a lot that contain i with same length.

			// location is Glasgow, which is near Wishaw
			lat: 55.8554403,
			lng: -4.3024976,
			check: func(t *testing.T, c []cities.CityWithScore) {
				// without distance, Wokingham is first result, but lat/lng is very far away
				// Wishaw is very close to Glasgow
				if c[0].Name != "Wishaw" {
					t.Fatalf("expected the first result to be Wishaw, but was %s", c[0].Name)
				}

				// Workington is in Cumbria, so comes next
				if c[1].Name != "Workington" {
					t.Fatalf("expected the second result to be Workington, but was %s", c[1].Name)
				}

			},
		},
		{
			name:  "very good match still trumps distance",
			csv:   citySample,
			query: "Wokin",
			// location is the Lincolnshire Wolds, very far from Woking
			lat: 53.3453018,
			lng: -0.2011261,
			check: func(t *testing.T, c []cities.CityWithScore) {
				// location is very far from woking, but we should still get it as first match
				// given it's a very strong string match
				if c[0].Name != "Woking" {
					t.Fatalf("expected the first result to be Woking, but was %s", c[0].Name)
				}
			},
		},
	}

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cs, err := cities.NewCitySearcher(strings.NewReader(tc.csv))
			if err != nil {
				t.Fatalf("failed to make city searcher: %s", err)
			}

			results, err := cs.SearchWithLocation(context.Background(), tc.query, tc.lat, tc.lng)
			if err != nil {
				t.Fatalf("failed to  search: %s", err)
			}

			fmt.Println(results)
			tc.check(t, results)
		})
	}
}
