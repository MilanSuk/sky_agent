package main

import (
	"log"
)

// Get user body measurement
type get_user_info struct {
	Item string // options["weight_kg", "height_meters", "age"]
}

func (st *get_user_info) run() float64 {
	switch st.Item {
	case "weight_kg":
		return 75
	case "height_meters":
		return 180
	case "age":
		return 30
	default:
		log.Fatalf("Invalid option: %s", st.Item)
	}
	return -1
}
