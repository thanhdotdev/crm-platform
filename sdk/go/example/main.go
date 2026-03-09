package main

import (
	"context"
	"log"
	"os"

	crmsdk "github.com/vothanh/crm-platform/sdk/go"
)

func main() {
	// Ensure CRM_API_KEY is available (passed from test_sdk.sh)
	if os.Getenv("CRM_API_KEY") == "" {
		log.Fatal("❌ CRM_API_KEY is not set. Please run via scripts/test_sdk.sh")
	}

	log.Println("Initializing CRM SDK Client...")
	client := crmsdk.NewClient()

	log.Println("Sending test event: 'app_opened'")
	err := client.Track(
		context.Background(),
		"app_opened",
		"test-user-001",
		"customer",
		map[string]interface{}{
			"device": "iphone_15",
			"os":     "ios_17",
		},
	)

	if err != nil {
		log.Fatalf("❌ Failed to send event via Go SDK: %v", err)
	}

	log.Println("✅ Successfully sent event via Go SDK!")
}
