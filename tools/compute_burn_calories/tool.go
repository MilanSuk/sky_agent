package main

import (
	"encoding/xml"
	"io"
	"log"
	"math"
	"os"
	"time"
)

// compute_burn_calories calculates calories burned from a GPX file.
type compute_burn_calories struct {
	FilePath string  // Path to the GPX file
	Weight   float64 // Weight of the user in kilograms
}

// Trkpt represents a track point in GPX file
type Trkpt struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Ele  float64 `xml:"ele"`
	Time string  `xml:"time"`
}

// Gpx represents the structure of a GPX file
type Gpx struct {
	XMLName xml.Name `xml:"gpx"`
	Trk     struct {
		Trkseg struct {
			Trkpt []Trkpt `xml:"trkpt"`
		} `xml:"trkseg"`
	} `xml:"trk"`
}

func (st *compute_burn_calories) run() float64 {
	gpxFile, err := os.Open(st.FilePath)
	if err != nil {
		log.Fatalf("Failed to open GPX file: %v", err)
	}
	defer gpxFile.Close()

	byteValue, _ := io.ReadAll(gpxFile)
	var gpxData Gpx
	xml.Unmarshal(byteValue, &gpxData)

	if len(gpxData.Trk.Trkseg.Trkpt) < 2 {
		log.Fatalf("Not enough track points to calculate distance and duration.")
	}

	totalDistance := 0.0
	startTime, _ := time.Parse(time.RFC3339, gpxData.Trk.Trkseg.Trkpt[0].Time)
	endTime, _ := time.Parse(time.RFC3339, gpxData.Trk.Trkseg.Trkpt[len(gpxData.Trk.Trkseg.Trkpt)-1].Time)

	for i := 0; i < len(gpxData.Trk.Trkseg.Trkpt)-1; i++ {
		point1 := gpxData.Trk.Trkseg.Trkpt[i]
		point2 := gpxData.Trk.Trkseg.Trkpt[i+1]
		totalDistance += haversine(point1.Lat, point1.Lon, point2.Lat, point2.Lon)
	}

	duration := endTime.Sub(startTime).Hours()
	caloriesBurned := calculateCalories(totalDistance, duration, st.Weight)

	return caloriesBurned
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // km

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	lat1 = toRadians(lat1)
	lat2 = toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}
func toRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func calculateCalories(distance_kms, duration_hours, weight_kgs float64) float64 {
	// Calculate speed in km/h
	speed := distance_kms / duration_hours

	// MET (Metabolic Equivalent of Task) varies based on speed
	// Using standard MET values for different walking/running speeds
	var met float64
	switch {
	case speed <= 4.0: // Slow walking
		met = 2.5
	case speed <= 6.0: // Moderate walking
		met = 3.5
	case speed <= 8.0: // Fast walking/slow jogging
		met = 6.0
	case speed <= 11.0: // Jogging
		met = 8.3
	default: // Running
		met = 11.0
	}

	// Calories = MET × weight (kg) × duration (hours)
	// The formula includes 1.05 as a constant to account for resting metabolism
	calories := met * weight_kgs * duration_hours * 1.05

	return calories
}
