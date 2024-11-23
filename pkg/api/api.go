package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
	"what-to-eat/pkg/structure"
)

type SimplifiedMenu struct {
	Name string         `json:"name"`
	Menu []MenuCategory `json:"menu"`
}

type MenuCategory struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Products    []Product `json:"products"`
}

type Product struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       float64     `json:"price"`
	ImageURL    string      `json:"image_url,omitempty"`
	Variations  []Variation `json:"variations,omitempty"`
}

type Variation struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func promoteAlgorithm(rating float64, reviewNumber int) float64 {
	maxReview := 200.0
	offset := 50.0
	weight := math.Min(1.0, maxReview/float64(reviewNumber))
	return (float64(reviewNumber)*rating + offset*weight) / (float64(reviewNumber) + offset*weight)
}

// cryptoRandInt generates a cryptographically secure random integer between 0 and max-1
func cryptoRandInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive")
	}

	bigInt := big.NewInt(int64(max))
	randInt, err := rand.Int(rand.Reader, bigInt)
	if err != nil {
		return 0, err
	}

	return int(randInt.Int64()), nil
}

func RandomRestaurantHandler(w http.ResponseWriter, r *http.Request) {
	// Get parameters from URL
	latStr := r.URL.Query().Get("latitude")
	lonStr := r.URL.Query().Get("longitude")
	cuisineTypesStr := r.URL.Query().Get("cuisineTypes")

	// Validate required parameters
	if latStr == "" || lonStr == "" {
		http.Error(w, "Missing latitude or longitude parameters", http.StatusBadRequest)
		return
	}

	// Convert string parameters to float64
	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude value: "+err.Error(), http.StatusBadRequest)
		return
	}
	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude value: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse cuisine types
	var cuisineTypes []string
	if cuisineTypesStr != "" {
		cuisineTypes = strings.Split(cuisineTypesStr, ",")
	}

	// Prepare Foodpanda API request
	client := &http.Client{
		Timeout: 10 * time.Second, // Add timeout for safety
	}
	foodpandaURL := "https://disco.deliveryhero.io/listing/api/v1/pandora/vendors"
	req, err := http.NewRequest("GET", foodpandaURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
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
	q.Add("limit", "999999")
	q.Add("offset", "0")
	q.Add("customer_type", "regular")

	// Add cuisine types if provided
	if len(cuisineTypes) > 0 {
		q.Add("cuisine", strings.Join(cuisineTypes, ","))
	}

	req.URL.RawQuery = q.Encode()
	req.Header.Add("x-disco-client-id", "web")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var foodpandaResp structure.FoodPandaRestaurantResponse
	if err := json.NewDecoder(resp.Body).Decode(&foodpandaResp); err != nil {
		http.Error(w, "Failed to parse response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Process restaurants and filter available ones
	var availableRestaurants []structure.Restaurant
	for _, item := range foodpandaResp.Data.Items {
		if item.Metadata.IsDeliveryAvailable {
			restaurant := structure.Restaurant{
				ID:                 item.ID,
				Name:               item.Name,
				Chain:              item.Chain,
				HeroImage:          item.HeroImage,
				Address:            item.Address,
				Distance:           item.Distance,
				Rating:             item.Rating,
				ReviewNumber:       item.ReviewNumber,
				RedirectionURL:     item.RedirectionURL,
				MinimumOrderAmount: item.MinimumOrderAmount,
				DeliveryFee:        item.MinimumDeliveryFee,
				DeliveryTime:       item.MinimumDeliveryTime,
				Weight:             promoteAlgorithm(item.Rating, item.ReviewNumber),
			}
			availableRestaurants = append(availableRestaurants, restaurant)
		}
	}

	// Check if we have any restaurants
	if len(availableRestaurants) == 0 {
		http.Error(w, "No available restaurants found", http.StatusNotFound)
		return
	}

	// Select a random restaurant using crypto/rand for better randomness
	randIndex, err := cryptoRandInt(len(availableRestaurants))
	if err != nil {
		http.Error(w, "Failed to generate random selection: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response with single random restaurant
	apiResponse := structure.ApiResponse{
		Restaurants: []structure.Restaurant{availableRestaurants[randIndex]},
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	json.NewEncoder(w).Encode(apiResponse)
}

func transformResponse(rawData []byte) ([]byte, error) {
	var fullResponse struct {
		Data struct {
			Name  string `json:"name"`
			Menus []struct {
				MenuCategories []struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					Products    []struct {
						Name              string `json:"name"`
						Description       string `json:"description"`
						FilePath          string `json:"file_path"`
						ProductVariations []struct {
							Name  string  `json:"name"`
							Price float64 `json:"price"`
						} `json:"product_variations"`
					} `json:"products"`
				} `json:"menu_categories"`
			} `json:"menus"`
		} `json:"data"`
	}

	if err := json.Unmarshal(rawData, &fullResponse); err != nil {
		return nil, err
	}

	// Create simplified menu structure
	simplified := SimplifiedMenu{
		Name: fullResponse.Data.Name,
	}

	// Process only the first menu (usually the main menu)
	if len(fullResponse.Data.Menus) > 0 {
		for _, category := range fullResponse.Data.Menus[0].MenuCategories {
			menuCat := MenuCategory{
				Name:        category.Name,
				Description: category.Description,
			}

			for _, prod := range category.Products {
				product := Product{
					Name:        prod.Name,
					Description: prod.Description,
					ImageURL:    prod.FilePath,
				}

				// Process variations
				for _, var_ := range prod.ProductVariations {
					product.Variations = append(product.Variations, Variation{
						Name:  var_.Name,
						Price: var_.Price,
					})
				}

				// If there's only one variation with no name, use it as the main price
				if len(product.Variations) == 1 && product.Variations[0].Name == "" {
					product.Price = product.Variations[0].Price
					product.Variations = nil
				}

				menuCat.Products = append(menuCat.Products, product)
			}

			simplified.Menu = append(simplified.Menu, menuCat)
		}
	}

	return json.Marshal(simplified)
}

// GetCuisinesHandler return near cuisine catogories
func GetCuisinesHandler(w http.ResponseWriter, r *http.Request) {
	// Get parameters from URL
	latStr := r.URL.Query().Get("latitude")
	lonStr := r.URL.Query().Get("longitude")

	// Validate parameters exist
	if latStr == "" || lonStr == "" {
		http.Error(w, "Missing latitude or longitude parameters", http.StatusBadRequest)
		return
	}

	// Convert string parameters to float64
	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude value: "+err.Error(), http.StatusBadRequest)
		return
	}

	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude value: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Prepare Foodpanda API request
	client := &http.Client{}
	foodpandaURL := "https://disco.deliveryhero.io/listing/api/v1/pandora/vendors"

	req, err := http.NewRequest("GET", foodpandaURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
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

	req.URL.RawQuery = q.Encode()
	req.Header.Add("x-disco-client-id", "web")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var foodpandaResp structure.FoodPandaRestaurantResponse
	if err := json.NewDecoder(resp.Body).Decode(&foodpandaResp); err != nil {
		http.Error(w, "Failed to parse response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the nested Cuisines struct to []CuisineInfo
	cuisinesInfo := make([]structure.CuisineInfo, len(foodpandaResp.Data.Aggregations.Cuisines))
	for i, cuisine := range foodpandaResp.Data.Aggregations.Cuisines {
		cuisinesInfo[i] = structure.CuisineInfo{
			ID:    cuisine.ID,
			Title: cuisine.Title,
			Count: cuisine.Count,
			Slug:  cuisine.Slug,
		}
	}

	// Create and send response
	cuisinesResponse := structure.CuisinesResponse{
		Cuisines: cuisinesInfo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cuisinesResponse)
}
