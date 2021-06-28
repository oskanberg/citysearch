package main

import (
	"net/http"

	"github.com/oskanberg/citysearch/api"
	"github.com/oskanberg/citysearch/cities"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	searcher, err := cities.NewCitySearcher()
	if err != nil {
		log.Fatalf("failed to create city searcher: %s", err)
	}

	// single endpoint, so don't feel the need to do any fancy muxing
	http.HandleFunc("/suggestions", api.NewCitySearchHandler(searcher))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
