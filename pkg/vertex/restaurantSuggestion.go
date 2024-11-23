package vertex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"what-to-eat/pkg/structure"
)

type RestaurantSuggestionRequestBody struct {
	InitialPreference string `json:"initial_preference"`
	AdditionalDetail  string `json:"additional_detail"`
	Location          struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	Cuisines []struct {
		ID    int    `json:"id"`
		Label string `json:"label"`
	} `json:"cuisines"`
}

type MenuFetchRestaurantInfo struct {
	Id             string `json:"id"`
	Heroimage      string `json:"hero_image"`
	Name           string `json:"name"`
	RedirectionURL string `json:"redirection_url"`
	Code           string `json:"code"`
	Longitude      string `json:"longitude"`
	Latitude       string `json:"latitude"`
}

type AiSuggestionRequestBody struct {
}

func init() {
	InitProjectInfo()
}

// fetchNearRestaurant fetches nearby restaurants from the Foodpanda API.
func fetchNearRestaurant(latitude float64, longitude float64, cuisineIDs []string) ([]MenuFetchRestaurantInfo, error) {
	// Prepare the Foodpanda API request
	client := &http.Client{
		Timeout: 10 * time.Second, // Add timeout for safety
	}
	foodpandaURL := "https://disco.deliveryhero.io/listing/api/v1/pandora/vendors"
	req, err := http.NewRequest("GET", foodpandaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Foodpanda API request: %w", err)
	}

	// Set query parameters
	q := req.URL.Query()
	q.Add("country", "tw")
	q.Add("latitude", fmt.Sprintf("%f", latitude))
	q.Add("longitude", fmt.Sprintf("%f", longitude))
	q.Add("language_id", "6")
	q.Add("include", "characteristics")
	q.Add("dynamic_pricing", "0")
	q.Add("configuration", "Original")
	q.Add("vertical", "restaurants")
	q.Add("limit", "1")
	q.Add("offset", "0")
	q.Add("customer_type", "regular")

	if len(cuisineIDs) > 0 {
		q.Add("cuisine", strings.Join(cuisineIDs, ","))
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Add("x-disco-client-id", "web")

	// Make the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making Foodpanda API request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the Foodpanda API response
	var foodpandaResp structure.FoodPandaRestaurantResponse
	if err := json.NewDecoder(resp.Body).Decode(&foodpandaResp); err != nil {
		return nil, fmt.Errorf("error decoding Foodpanda API response: %w", err)
	}
	fmt.Println("Foodpanda resp", foodpandaResp.Data.Items)

	// Transform response into MenuFetchRestaurantInfo format
	var restaurantInfos []MenuFetchRestaurantInfo
	for _, item := range foodpandaResp.Data.Items {
		if item.Metadata.IsDeliveryAvailable {
			info := MenuFetchRestaurantInfo{
				Id:             fmt.Sprintf("%d", item.ID), // Convert int to string
				Heroimage:      item.HeroImage,
				Name:           item.Name,
				RedirectionURL: item.RedirectionURL,
				Code:           item.Code,
			}
			restaurantInfos = append(restaurantInfos, info)
		}
	}
	fmt.Println("Restaurant info", restaurantInfos)

	return restaurantInfos, nil
}

func fetchRestaurantMenu(restaurantCodes []string, latitude, longitude float64) (map[string][]byte, error) {
	// Prepare the HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second, // Add timeout for safety
	}

	// Prepare the request payload in the new structure
	payload := map[string]interface{}{
		"code":      restaurantCodes,
		"longitude": longitude,
		"latitude":  latitude,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Make the POST request to localhost:3001/menu
	req, err := http.NewRequest("POST", "http://localhost:3001/menu", strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create menu request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch restaurant menus: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("menu request returned non-OK status: %d", resp.StatusCode)
	}

	// Parse the response
	var menuResponse struct {
		Success bool                     `json:"success"`
		Results []map[string]interface{} `json:"results"` // Adjusted based on expected structure
	}
	if err := json.NewDecoder(resp.Body).Decode(&menuResponse); err != nil {
		return nil, fmt.Errorf("failed to decode menu response: %w", err)
	}

	// Check if the success flag is false
	if !menuResponse.Success {
		return nil, fmt.Errorf("menu request failed")
	}

	// Map results for easier handling
	menuMap := make(map[string][]byte)
	for _, result := range menuResponse.Results {
		code := result["code"].(string) // Ensure safe casting
		menuJSON, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal menu result: %w", err)
		}
		menuMap[code] = menuJSON
	}

	return menuMap, nil
}

// func aiSuggestion() {

// }

func RestaurantSuggestion(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var requestBody RestaurantSuggestionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		fmt.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract location data
	latitude := requestBody.Location.Latitude
	longitude := requestBody.Location.Longitude

	// Extract cuisine IDs
	var cuisineIDs []string
	for _, cuisine := range requestBody.Cuisines {
		cuisineIDs = append(cuisineIDs, fmt.Sprintf("%d", cuisine.ID))
	}

	// Fetch nearby restaurants
	restaurantInfos, err := fetchNearRestaurant(latitude, longitude, cuisineIDs)
	if err != nil {
		fmt.Println("Error fetching restaurants:", err)
		http.Error(w, "Failed to fetch restaurants: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Collect restaurant IDs
	var restaurantCodes []string
	for _, info := range restaurantInfos {
		restaurantCodes = append(restaurantCodes, info.Code)
	}

	// Fetch menus for all restaurants
	menus, err := fetchRestaurantMenu(restaurantCodes, latitude, longitude) // Pass latitude and longitude
	if err != nil {
		fmt.Println("Error fetching menus:", err)
		http.Error(w, "Failed to fetch menus: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the final response
	response := struct {
		Restaurants []MenuFetchRestaurantInfo `json:"restaurants"`
		Menus       map[string][]byte         `json:"menus"`
	}{
		Restaurants: restaurantInfos,
		Menus:       menus,
	}

	// Return the final response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Println("Error encoding response:", err)
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
