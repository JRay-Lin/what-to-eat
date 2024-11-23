package structure

// FoodPandaResponse represents the root response from Foodpanda API
type FoodPandaRestaurantResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Data       struct {
		AvailableCount int              `json:"available_count"`
		ReturnedCount  int              `json:"returned_count"`
		Items          []RestaurantItem `json:"items"`
		Aggregations   AggregationsData `json:"aggregations"`
	} `json:"data"`
}
