package structure

import (
	"time"
)

type FoodPandaMenuResponse struct {
	StatusCode int `json:"status_code"`
	Data       struct {
		Menus []FoodPandaMenu `json:"menus"`
	} `json:"data"`
}

type FoodPandaMenu struct {
	ID                    int                `json:"id"`
	Code                  string             `json:"code"`
	Name                  string             `json:"name"`
	Description           string             `json:"description"`
	Type                  string             `json:"type"`
	OpeningTime           string             `json:"opening_time"`
	ClosingTime           string             `json:"closing_time"`
	MenuCategories        []MenuCategory     `json:"menu_categories"`
	Toppings              map[string]Topping `json:"toppings"`
	Tags                  map[string]MenuTag `json:"tags"`
	DefaultSoldOutOptions []SoldOutOption    `json:"default_sold_out_options"`
}

type MenuCategory struct {
	ID                int       `json:"id"`
	Code              string    `json:"code"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Products          []Product `json:"products"`
	IsPopularCategory bool      `json:"is_popular_category"`
}

type Product struct {
	ID                int                `json:"id"`
	Code              string             `json:"code"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	MasterCategoryID  int                `json:"master_category_id"`
	FilePath          string             `json:"file_path"`
	IsSoldOut         bool               `json:"is_sold_out"`
	IsExpressItem     bool               `json:"is_express_item"`
	ProductVariations []ProductVariation `json:"product_variations"`
	Tags              []string           `json:"tags,omitempty"`
	IsBundle          bool               `json:"is_bundle"`
}

type ProductVariation struct {
	ID             int         `json:"id"`
	Code           string      `json:"code"`
	RemoteCode     string      `json:"remote_code"`
	ContainerPrice int         `json:"container_price"`
	Name           string      `json:"name,omitempty"`
	Price          int         `json:"price"`
	ToppingIDs     []int       `json:"topping_ids"`
	UnitPricing    interface{} `json:"unit_pricing"`
	TotalPrice     int         `json:"total_price"`
}

type Topping struct {
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	QuantityMinimum int             `json:"quantity_minimum"`
	QuantityMaximum int             `json:"quantity_maximum"`
	Options         []ToppingOption `json:"options"`
	Type            string          `json:"type"`
}

type ToppingOption struct {
	ID          int    `json:"id"`
	ProductID   int    `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	RemoteCode  string `json:"remote_code"`
}

type MenuTag struct {
	Name            string            `json:"name"`
	TranslationKeys map[string]string `json:"translation_keys"`
	Elements        []string          `json:"elements"`
	Metadata        TagMetadata       `json:"metadata"`
}

type TagMetadata struct {
	Sorting []int `json:"sorting"`
}

type SoldOutOption struct {
	Default bool   `json:"default"`
	Option  string `json:"option"`
	Text    string `json:"text"`
}

// RestaurantItem represents a single restaurant from the response
type RestaurantItem struct {
	ID      int    `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Budget  int    `json:"budget"`
	Chain   struct {
		Code           string `json:"code"`
		Name           string `json:"name"`
		MainVendorCode string `json:"main_vendor_code"`
		URLKey         string `json:"url_key"`
	} `json:"chain"`
	Cuisines []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		URLKey string `json:"url_key"`
		Main   bool   `json:"main"`
	} `json:"cuisines"`
	Distance       float64 `json:"distance"`
	HeroImage      string  `json:"hero_image"`
	Rating         float64 `json:"rating"`
	ReviewNumber   int     `json:"review_number"`
	RedirectionURL string  `json:"redirection_url"`
	Metadata       struct {
		HasDiscount         bool `json:"has_discount"`
		IsDeliveryAvailable bool `json:"is_delivery_available"`
		IsPickupAvailable   bool `json:"is_pickup_available"`
		IsTemporaryClosed   bool `json:"is_temporary_closed"`
	} `json:"metadata"`
	MinimumDeliveryFee  float64 `json:"minimum_delivery_fee"`
	MinimumDeliveryTime float64 `json:"minimum_delivery_time"`
	MinimumOrderAmount  float64 `json:"minimum_order_amount"`
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
}

// AggregationsData represents the aggregations section of the response
type AggregationsData struct {
	Cuisines []struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Count int    `json:"count"`
		Slug  string `json:"slug"`
	} `json:"cuisines"`
}

// RequestParams represents the parameters sent from the client
type RequestParams struct {
	Location  Location `json:"location"`
	FoodTypes []string `json:"foodTypes"`
}

// Location represents geographical coordinates
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Restaurant represents our processed restaurant data for response
type Restaurant struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Chain struct {
		Code           string `json:"code"`
		Name           string `json:"name"`
		MainVendorCode string `json:"main_vendor_code"`
		URLKey         string `json:"url_key"`
	} `json:"chain"`
	HeroImage          string  `json:"hero_image"`
	Address            string  `json:"address"`
	Distance           float64 `json:"distance"`
	Rating             float64 `json:"rating"`
	ReviewNumber       int     `json:"review_number"`
	RedirectionURL     string  `json:"redirection_url"`
	MinimumOrderAmount float64 `json:"minimum_order_amount"`
	DeliveryFee        float64 `json:"delivery_fee"`
	DeliveryTime       float64 `json:"delivery_time"`
	Weight             float64 `json:"weight"`
}

// ApiResponse represents our API's response structure
type ApiResponse struct {
	Restaurants []Restaurant       `json:"restaurants"`
	Cuisines    []AggregationsData `json:"cuisines"`
}

type CuisineInfo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Count int    `json:"count"`
	Slug  string `json:"slug"`
}

type PerimeterConfig struct {
	Timeout    time.Duration
	BaseURL    string
	MaxRetries int
	Debug      bool
}

type CuisinesResponse struct {
	Cuisines []CuisineInfo `json:"cuisines"`
}
