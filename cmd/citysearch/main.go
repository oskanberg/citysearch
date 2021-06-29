package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/oskanberg/citysearch/api"
	"github.com/oskanberg/citysearch/cities"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)

	fLoc := flag.String("cities", "", "location of the cities csv file")
	fPort := flag.String("port", ":80", "port to serve on")
	flag.Parse()

	if fLoc == nil || *fLoc == "" {
		log.Fatalf("flag --cities must be set to the location of the cities database")
	}

	f, err := os.Open(*fLoc)
	if err != nil {
		log.Fatalf("cities database could not be opened: %s", err)
	}

	searcher, err := cities.NewCitySearcher(f, cities.OnlyGB)
	if err != nil {
		log.Fatalf("failed to create city searcher: %s", err)
	}

	// single endpoint, so don't feel the need to do any fancy muxing
	http.HandleFunc("/suggestions", api.NewCitySearchHandler(searcher))

	log.Info("Service starting on port ", *fPort)
	log.Fatal(http.ListenAndServe(*fPort, nil))
}
