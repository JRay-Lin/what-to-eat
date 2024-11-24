package vertex

import (
	"encoding/json"
	"fmt"
	"io"
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

type GeminiSuggestionRequestBody struct {
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
	Menus map[string]interface{}
}

type GeminiSuggestionRespond struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
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
		Timeout: 25 * time.Second, // Add timeout for safety
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
	q.Add("limit", "999")
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
	// fmt.Println("Foodpanda resp", foodpandaResp.Data.Items)

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
				Longitude:      fmt.Sprintf("%f", longitude),
				Latitude:       fmt.Sprintf("%f", latitude),
			}
			restaurantInfos = append(restaurantInfos, info)
		}
	}
	// fmt.Println("Restaurant info", restaurantInfos)

	return restaurantInfos, nil
}

func fetchRestaurantMenu(restaurantCodes []string, latitude, longitude float64) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload := map[string]interface{}{
		"code":      restaurantCodes,
		"longitude": longitude,
		"latitude":  latitude,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:3001/menu", strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create menu request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch restaurant menus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("menu request returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
	}

	var menuResponse struct {
		Success bool                     `json:"success"`
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&menuResponse); err != nil {
		return nil, fmt.Errorf("failed to decode menu response: %w", err)
	}

	if !menuResponse.Success {
		return nil, fmt.Errorf("menu request failed")
	}

	menuMap := make(map[string]interface{})
	for _, result := range menuResponse.Results {
		code, ok := result["code"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid code type")
		}
		menuMap[code] = result // Keep as raw interface{}
	}

	return menuMap, nil
}

func aiSuggestion(requestBody RestaurantSuggestionRequestBody, menus map[string]interface{}) (GeminiSuggestionRespond, error) {
	// // Combine request body and menus into a single struct
	// aiRequestBody := GeminiSuggestionRequestBody{
	// 	InitialPreference: requestBody.InitialPreference,
	// 	AdditionalDetail:  requestBody.AdditionalDetail,
	// 	Location:          requestBody.Location,
	// 	Cuisines:          requestBody.Cuisines,
	// 	Menus:             menus,
	// }

	// // Convert the struct to JSON
	// aiRequestBodyJSON, err := json.Marshal(aiRequestBody)
	// if err != nil {
	// 	return GeminiSuggestionRespond{}, fmt.Errorf("failed to marshal AI request body: %w", err)
	// }

	// // Create AI request payload
	// requestPayload := structure.VertexAIRequest{
	// 	Contents: []struct {
	// 		Role  string `json:"role"`
	// 		Parts []struct {
	// 			Text string `json:"text"`
	// 		} `json:"parts"`
	// 	}{
	// 		{
	// 			Role: "user",
	// 			Parts: []struct {
	// 				Text string `json:"text"`
	// 			}{
	// 				{
	// 					Text: string(aiRequestBodyJSON),
	// 				},
	// 			},
	// 		},
	// 	},

	// 	SystemInstruction: struct {
	// 		Parts []struct {
	// 			Text string `json:"text"`
	// 		} `json:"parts"`
	// 	}{
	// 		Parts: []struct {
	// 			Text string `json:"text"`
	// 		}{
	// 			{
	// 				Text: "You are a restaurant picker bot. You will receive user preferences and local restaurant data in JSON format. Your task is to analyze this data and determine the restaurant that best fits the user's needs.\n\nYou will receive input in the following JSON structure:\n\n```json\n{\n    \"initial_preference\": \"user's initial preference\",\n    \"additional_Details\": \"additional details about user's preference\",\n    \"location\": {\n        \"latitude\": \"latitude of user's location\",\n        \"longitude\": \"longitude of user's location\"\n    },\n    \"cuisines\": [\n        {\n            \"id\": \"cuisine ID\",\n            \"label\": \"cuisine label\"\n        },\n        // ... more cuisines\n    ],\n    \"menus\": {\n        \"restaurant_code_1\": {\n            \"code\": \"restaurant code\",\n            \"name\": \"restaurant name\",\n            \"web_path\": \"restaurant web path\",\n            \"menus\": [\n                {\n                    \"id\": \"menu ID\",\n                    \"menu_categories\": [\n                        {\n                            \"description\": \"category description\",\n                            \"id\": \"category ID\",\n                            \"name\": \"category name\",\n                            \"menu_items\": [\n                                {\n                                    \"description\": \"item description\",\n                                    \"id\": \"item ID\",\n                                    \"name\": \"item name\",\n                                    \"price\": \"item price\"\n                                },\n                                // ... more menu items\n                            ]\n                        },\n                        // ... more menu categories\n                    ]\n                },\n                // ... more menus\n            ]\n        },\n        // ... more restaurants\n    }\n}\n```\n\nBased on this information, determine the restaurant that best suits the user's preferences, considering their initial preference, additional details, location, preferred cuisines, and the available menu items.  Pay close attention to the user's desired spice level.\n\nGenerate a JSON response in the following format:\n\n```json\n{\n\t\"code\":\"restaurant_code\",\n\t\"reason\":\"your reason for choosing this restaurant\"\n}\n```\n\nEnsure the `reason` field clearly and concisely explains why the chosen restaurant is the best match for the user.  Consider all available information when making your decision.",
	// 			},
	// 		},
	// 	},
	// 	GenerationConfig: struct {
	// 		Temperature     float64 `json:"temperature"`
	// 		MaxOutputTokens int     `json:"maxOutputTokens"`
	// 		TopP            float64 `json:"topP"`
	// 		Seed            int     `json:"seed"`
	// 	}{
	// 		Temperature:     0.1,
	// 		MaxOutputTokens: 8192,
	// 		TopP:            0.95,
	// 		Seed:            0,
	// 	},
	// 	SafetySettings: []struct {
	// 		Category  string `json:"category"`
	// 		Threshold string `json:"threshold"`
	// 	}{
	// 		{
	// 			Category:  "HARM_CATEGORY_HATE_SPEECH",
	// 			Threshold: "OFF",
	// 		},
	// 	},
	// }

	// payloadBytes, err := json.Marshal(requestPayload)
	// if err != nil {
	// 	return GeminiSuggestionRespond{}, fmt.Errorf("failed to marshal payload: %w", err)
	// }

	// // Send the request to Vertex AI
	// apiURL := fmt.Sprintf(
	// 	"https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent",
	// 	ProjectInfo.ApiEndpoint, ProjectInfo.ProjectID, ProjectInfo.Location, ProjectInfo.ModelID,
	// )
	// req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	// if err != nil {
	// 	return GeminiSuggestionRespond{}, fmt.Errorf("failed to create request: %w", err)
	// }

	// // Get google oauth token
	// accessToken, err := OauthGoogle()
	// if err != nil {
	// 	return GeminiSuggestionRespond{}, fmt.Errorf("failed to get access token: %w", err)
	// }

	// // Set headers
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// // Execute the request
	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return GeminiSuggestionRespond{}, fmt.Errorf("failed to send request: %w", err)
	// }
	// defer resp.Body.Close()

	// // Read the response
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return GeminiSuggestionRespond{}, fmt.Errorf("failed to read response body: %w", err)
	// }

	// // Parse the response
	// parsedResponse, err := ParseVertexAIResponse(string(body))
	// // fmt.Println(parsedResponse)

	// // Create an instance of the struct
	// var aiResponse GeminiSuggestionRespond
	// unmarshalErr := json.Unmarshal([]byte(parsedResponse), &aiResponse)
	// if unmarshalErr != nil {
	// 	return GeminiSuggestionRespond{}, fmt.Errorf("failed to unmarshal AI response: %w", err)
	// }

	// DEBUG
	var firstCode string
	for _, value := range menus {
		menuMap := value.(map[string]interface{})
		if code, ok := menuMap["code"].(string); ok {
			firstCode = code
			fmt.Println("First Code Found:", firstCode)
			break
		}
	}
	aiResponse := GeminiSuggestionRespond{
		Code:   firstCode,
		Reason: "test",
	}

	return aiResponse, nil
}

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
	menus, err := fetchRestaurantMenu(restaurantCodes, latitude, longitude)
	if err != nil {
		fmt.Println("Error fetching menus:", err)
		http.Error(w, "Failed to fetch menus: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Println("Menus:", menus)

	// Send menus and user preference to AI
	suggestion, err := aiSuggestion(requestBody, menus)
	if err != nil {
		fmt.Println("Error getting AI suggestion:", err)
		http.Error(w, "Failed to get AI suggestion: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Find restaurant that match the suggestion.Code
	var matchedRestaurant MenuFetchRestaurantInfo
	found := false

	for _, restaurant := range restaurantInfos {
		if restaurant.Code == suggestion.Code {
			matchedRestaurant = restaurant
			found = true
			break
		}
	}

	if found {
		fmt.Printf("Matched Restaurant: %+v\n", matchedRestaurant)
	} else {
		fmt.Printf("No restaurant matches the code: %s\n", suggestion.Code)
		// fmt.Println(restaurantInfos)
	}

	// Return the final response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matchedRestaurant); err != nil {
		fmt.Println("Error encoding response:", err)
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
