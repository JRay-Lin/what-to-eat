package vertex

import (
	"context"
	"log"
	"what-to-eat/pkg/structure"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var ProjectInfo structure.GeminiProjectInfo
var tokenSource oauth2.TokenSource

// InitProjectInfo initializes the project information
func InitProjectInfo() {
	ProjectInfo = structure.GeminiProjectInfo{
		ProjectID:   "what-to-eat-442102",
		Location:    "us-central1",
		ModelID:     "gemini-1.5-flash-002",
		ApiEndpoint: "us-central1-aiplatform.googleapis.com",
	}
}

func OauthGoogle() (string, error) {
	if tokenSource == nil {
		// Create a context
		ctx := context.Background()

		// Use the default Google credentials
		creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			log.Printf("Failed to get default credentials: %v", err)
			return "", err
		}

		// Create a token source
		tokenSource = creds.TokenSource
	}

	// Get a new token
	token, err := tokenSource.Token()
	if err != nil {
		log.Printf("Failed to retrieve token: %v", err)
		return "", err
	}

	return token.AccessToken, nil
}
