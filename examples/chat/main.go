package main

import (
	"context"
	"fmt"
	"os"

	"github.com/muxi-ai/muxi-go"
)

// Simple chat (non-streaming)
func main() {
	ctx := context.Background()

	client := muxi.NewFormationClient(&muxi.FormationConfig{
		FormationID: "my-bot",
		ServerURL:   os.Getenv("MUXI_SERVER_URL"), // proxy mode
		ClientKey:   os.Getenv("MUXI_CLIENT_KEY"),
		MaxRetries:  2,
	})

	resp, err := client.Chat(ctx, &muxi.ChatRequest{
		Message: "Hello, MUXI!",
		UserID:  "user-123",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response: %s\n", resp.Response)
	fmt.Printf("Request ID: %s\n", resp.RequestID)
}
