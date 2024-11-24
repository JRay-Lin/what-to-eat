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

		SystemInstruction: struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			Parts: []struct {
				Text string `json:"text"`
			}{
				{
					Text: "You are a culinary consultant specializing in recommending cuisines based on user preferences. You will receive a JSON object containing user preferences, location data, and a list of available cuisine categories. Your task is to analyze the user preferences and select the categories that best match those preferences.  Return the selected categories in a JSON array of objects, where each object contains the 'id' and 'label' of the selected category.  If no categories match the user's preferences, return an empty JSON array.\n\nInput JSON:\n```json\n{\"userPreference\": \"User's preferences \", \"location\": {\"latitude\": 24.1779755, \"longitude\": 120.6494471}, \"availableCategories\": [{\"id\": 163, \"label\": \"三明治 / 吐司\"}, {\"id\": 166, \"label\": \"中式\"}, {\"id\": 1210, \"label\": \"丼飯/蓋飯\"}, {\"id\": 1215, \"label\": \"便當\"}, {\"id\": 225, \"label\": \"健康餐\"}, {\"id\": 248, \"label\": \"台式\"}, {\"id\": 1212, \"label\": \"咖哩\"}, {\"id\": 1206, \"label\": \"咖啡\"}, {\"id\": 180, \"label\": \"壽司\"}, {\"id\": 214, \"label\": \"小吃\"}, {\"id\": 165, \"label\": \"披薩\"}, {\"id\": 1203, \"label\": \"拉麵\"}, {\"id\": 164, \"label\": \"日式\"}, {\"id\": 198, \"label\": \"早餐\"}, {\"id\": 252, \"label\": \"東南亞\"}, {\"id\": 179, \"label\": \"歐美\"}, {\"id\": 168, \"label\": \"泰式\"}, {\"id\": 235, \"label\": \"港式\"}, {\"id\": 199, \"label\": \"湯品\"}, {\"id\": 1220, \"label\": \"滷味\"}, {\"id\": 177, \"label\": \"漢堡\"}, {\"id\": 1214, \"label\": \"火鍋\"}, {\"id\": 1227, \"label\": \"炒飯\"}, {\"id\": 1209, \"label\": \"炸雞\"}, {\"id\": 1236, \"label\": \"燒烤\"}, {\"id\": 1211, \"label\": \"牛排\"}, {\"id\": 1241, \"label\": \"甜甜圈\"}, {\"id\": 176, \"label\": \"甜點\"}, {\"id\": 171, \"label\": \"異國\"}, {\"id\": 1202, \"label\": \"粥\"}, {\"id\": 186, \"label\": \"素食\"}, {\"id\": 195, \"label\": \"義大利麵\"}, {\"id\": 1216, \"label\": \"蛋糕\"}, {\"id\": 1233, \"label\": \"豆花\"}, {\"id\": 193, \"label\": \"越式\"}, {\"id\": 189, \"label\": \"鐵板燒\"}, {\"id\": 188, \"label\": \"韓式\"}, {\"id\": 181, \"label\": \"飲料\"}, {\"id\": 1208, \"label\": \"餃子\"}, {\"id\": 1221, \"label\": \"鹹酥雞/雞排\"}, {\"id\": 201, \"label\": \"麵食\"}]}\n```\n\nOutput JSON:\n```json\n[{\"id\": ..., \"label\": \"...\"}, {\"id\": ..., \"label\": \"...\"}, ...]\n```",
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
