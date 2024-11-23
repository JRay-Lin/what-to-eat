package vertex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"what-to-eat/pkg/structure"
)

type FilteredCategoriesRequestBody struct {
	UserPreference string `json:"userPreference"`
	Location       struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	AvailableCategories []struct {
		ID    int    `json:"id"`
		Label string `json:"label"`
	} `json:"availableCategories"`
}

func init() {
	InitProjectInfo()
}

func FilteredCategories(w http.ResponseWriter, r *http.Request) {
	var requestBody FilteredCategoriesRequestBody

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	// Convert the request body to the format expected by the AI
	aiInput, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal AI input: %v", err), http.StatusInternalServerError)
		return
	}

	// Create AI request payload
	requestPayload := structure.VertexAIRequest{
		Contents: []struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Role: "user",
				Parts: []struct {
					Text string `json:"text"`
				}{
					{
						Text: string(aiInput),
					},
				},
			},
		},
		// Rest of the configuration remains the same
		SystemInstruction: struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			Parts: []struct {
				Text string `json:"text"`
			}{
				{
					Text: "You are a culinary consultant specializing in recommending cuisines based on user preferences. You will receive a JSON object containing user preferences, location data, and a list of available cuisine categories. Your task is to analyze the user preferences and select the categories that best match those preferences. Return the selected categories in a JSON array of objects, where each object contains the 'id' and 'label' of the selected category. If no categories match the user's preferences, return an empty JSON array.",
				},
			},
		},
		GenerationConfig: struct {
			Temperature     float64 `json:"temperature"`
			MaxOutputTokens int     `json:"maxOutputTokens"`
			TopP            float64 `json:"topP"`
			Seed            int     `json:"seed"`
		}{
			Temperature:     0.1,
			MaxOutputTokens: 8192,
			TopP:            0.95,
			Seed:            0,
		},
		SafetySettings: []struct {
			Category  string `json:"category"`
			Threshold string `json:"threshold"`
		}{
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: "OFF",
			},
		},
	}

	// Convert to JSON
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal payload: %v", err), http.StatusInternalServerError)
		return
	}

	// Send the request to Vertex AI
	apiURL := fmt.Sprintf(
		"https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent",
		ProjectInfo.ApiEndpoint, ProjectInfo.ProjectID, ProjectInfo.Location, ProjectInfo.ModelID,
	)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	// Get google oauth token
	accessToken, err := OauthGoogle()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate access token: %v", err), http.StatusInternalServerError)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse the response
	parsedResponse, err := ParseVertexAIResponse(string(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse response: %v", err), http.StatusInternalServerError)
		return
	}

	// Send the parsed response back to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(parsedResponse))
}
