package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"
)

// Coordinate represents a geographic point
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// RouteRequest represents a route request from coordinates.json
type RouteRequest struct {
	ID   string     `json:"id"`
	From Coordinate `json:"from"`
	To   Coordinate `json:"to"`
}

// GeoJSONPoint represents a GeoJSON Feature with Point geometry
type GeoJSONPoint struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   struct {
		Type        string     `json:"type"`
		Coordinates [2]float64 `json:"coordinates"`
	} `json:"geometry"`
}

// GeoJSONFeatureCollection represents a GeoJSON FeatureCollection
type GeoJSONFeatureCollection struct {
	Type     string         `json:"type"`
	Features []GeoJSONPoint `json:"features"`
}

// RouteResult represents the combined route data
type RouteResult struct {
	ID            string                 `json:"id"`
	From          Coordinate             `json:"from"`
	To            Coordinate             `json:"to"`
	Distance      float64                `json:"distance_meters"`
	Time          float64                `json:"time_seconds"`
	RouteResponse map[string]interface{} `json:"route_response"`
	Error         string                 `json:"error,omitempty"`
}

const (
	graphhopperURL = "http://localhost:8989/route"
	profile        = "car"
)

func main() {
	// Read coordinates from GeoJSON file
	coordinates, err := readGeoJSON("input/Coordinates_for_routes.geojson")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading coordinates: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Processing %d routes...\n", len(coordinates))

	// Process each route request
	results := make([]RouteResult, 0, len(coordinates))
	for i, coord := range coordinates {
		fmt.Printf("[%d/%d] Computing route for %s...\n", i+1, len(coordinates), coord.ID)

		result := processRoute(coord)
		results = append(results, result)

		// Small delay to avoid overwhelming the API
		if i < len(coordinates)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Write results to JSON file
	if err := writeResults("output/route_results.json", results); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing results: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully processed %d routes. Results saved to output/route_results.json\n", len(results))
}

func readGeoJSON(filename string) ([]RouteRequest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var featureCollection GeoJSONFeatureCollection
	if err := json.Unmarshal(data, &featureCollection); err != nil {
		return nil, fmt.Errorf("failed to parse GeoJSON: %w", err)
	}

	// Group features by pair_id
	pairs := make(map[int]map[string]*GeoJSONPoint)
	for i := range featureCollection.Features {
		feature := &featureCollection.Features[i]

		// Extract pair_id from properties
		pairID, ok := feature.Properties["pair_id"].(float64)
		if !ok {
			continue
		}
		role, ok := feature.Properties["role"].(string)
		if !ok {
			continue
		}

		if _, exists := pairs[int(pairID)]; !exists {
			pairs[int(pairID)] = make(map[string]*GeoJSONPoint)
		}
		pairs[int(pairID)][role] = feature
	}

	// Create route requests from pairs
	var routes []RouteRequest
	var pairIDs []int
	for pairID := range pairs {
		pairIDs = append(pairIDs, pairID)
	}
	sort.Ints(pairIDs)

	for _, pairID := range pairIDs {
		pair := pairs[pairID]
		origin, hasOrigin := pair["origin"]
		destination, hasDestination := pair["destination"]

		if !hasOrigin || !hasDestination {
			fmt.Fprintf(os.Stderr, "Warning: pair_id %d missing origin or destination\n", pairID)
			continue
		}

		// Extract coordinates from geometry
		fromCoord := Coordinate{
			Lat: origin.Geometry.Coordinates[1],
			Lon: origin.Geometry.Coordinates[0],
		}
		toCoord := Coordinate{
			Lat: destination.Geometry.Coordinates[1],
			Lon: destination.Geometry.Coordinates[0],
		}

		route := RouteRequest{
			ID:   fmt.Sprintf("pair_%d", pairID),
			From: fromCoord,
			To:   toCoord,
		}
		routes = append(routes, route)
	}

	return routes, nil
}

func processRoute(req RouteRequest) RouteResult {
	result := RouteResult{
		ID:   req.ID,
		From: req.From,
		To:   req.To,
	}

	// Build GraphHopper API request
	apiURL, err := buildGraphHopperURL(req)
	if err != nil {
		result.Error = fmt.Sprintf("failed to build URL: %v", err)
		return result
	}

	// Make HTTP request
	resp, err := http.Get(apiURL)
	if err != nil {
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read response: %v", err)
		return result
	}

	// Parse response
	var routeData map[string]interface{}
	if err := json.Unmarshal(body, &routeData); err != nil {
		result.Error = fmt.Sprintf("failed to parse response: %v", err)
		return result
	}

	// Check for GraphHopper errors
	if msg, ok := routeData["message"].(string); ok && msg != "" {
		result.Error = msg
		result.RouteResponse = routeData
		return result
	}

	// Extract distance and time from the first path
	if paths, ok := routeData["paths"].([]interface{}); ok && len(paths) > 0 {
		if path, ok := paths[0].(map[string]interface{}); ok {
			if distance, ok := path["distance"].(float64); ok {
				result.Distance = distance
			}
			if timeMs, ok := path["time"].(float64); ok {
				result.Time = timeMs / 1000.0 // Convert milliseconds to seconds
			}
		}
	}

	result.RouteResponse = routeData
	return result
}

func buildGraphHopperURL(req RouteRequest) (string, error) {
	params := url.Values{}
	params.Add("point", fmt.Sprintf("%f,%f", req.From.Lat, req.From.Lon))
	params.Add("point", fmt.Sprintf("%f,%f", req.To.Lat, req.To.Lon))
	params.Add("profile", profile)
	params.Add("locale", "en")
	params.Add("points_encoded", "false")

	return fmt.Sprintf("%s?%s", graphhopperURL, params.Encode()), nil
}

func writeResults(filename string, results []RouteResult) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
