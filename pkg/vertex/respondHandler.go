package vertex

import (
	"encoding/json"
	"strings"
)

func ParseVertexAIResponse(responseData string) (string, error) {
	var result strings.Builder
	var responses []map[string]interface{}
	if err := json.Unmarshal([]byte(responseData), &responses); err != nil {
		return "", err
	}

	for _, response := range responses {
		if candidates, ok := response["candidates"].([]interface{}); ok {
			for _, candidate := range candidates {
				if content, ok := candidate.(map[string]interface{})["content"].(map[string]interface{}); ok {
					if parts, ok := content["parts"].([]interface{}); ok {
						for _, part := range parts {
							if text, ok := part.(map[string]interface{})["text"].(string); ok {
								result.WriteString(text)
							}
						}
					}
				}
			}
		}
	}

	response := result.String()
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimPrefix(response, "json\n")
	response = strings.TrimSuffix(response, "\n```\n")

	return response, nil
}
