package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Get the population of a given city. Using OpenStreetMap's Nominatim API.
type get_city_population struct {
	City string //The city name.
}

type OSMResponse struct {
	Place_id     int64   `json:"place_id"`
	Licence      string  `json:"licence"`
	OSM_type     string  `json:"osm_type"`
	OSM_id       int64   `json:"osm_id"`
	Lat          string  `json:"lat"`
	Lon          string  `json:"lon"`
	Display_name string  `json:"display_name"`
	Address      Address `json:"address"`
	ExtratTags   Tags    `json:"extratags"`
}

type Address struct {
	City         string `json:"city"`
	Country      string `json:"country"`
	Country_code string `json:"country_code"`
}

type Tags struct {
	Population string `json:"population"`
}

func (st *get_city_population) run() int64 {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build the URL with proper encoding
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Add("q", st.City)
	params.Add("format", "json")
	params.Add("addressdetails", "1")
	params.Add("extratags", "1")
	params.Add("limit", "1")

	// Add required headers
	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		log.Fatalf("error creating request: %v", err)
	}

	// Add User-Agent header as required by OSM API
	req.Header.Set("User-Agent", "CityPopulationTool/1.0")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response: %v", err)
	}

	// Parse the JSON response
	var osmResponse []OSMResponse
	if err := json.Unmarshal(body, &osmResponse); err != nil {
		log.Fatalf("error parsing JSON: %v", err)
	}

	// Check if we got any results
	if len(osmResponse) == 0 {
		log.Fatalf("no results found for city: %s", st.City)
	}

	// Try to parse the population
	var population int64
	if osmResponse[0].ExtratTags.Population != "" {
		fmt.Sscanf(osmResponse[0].ExtratTags.Population, "%d", &population)
	}

	if population == 0 {
		log.Fatalf("population data not available for city: %s", st.City)
	}

	return population
}
