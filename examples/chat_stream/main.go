package main

import (
	"context"
	"fmt"
	"os"

	"github.com/muxi-ai/muxi-go"
)

// Chat streaming example
func main() {
	ctx := context.Background()

	client := muxi.NewFormationClient(&muxi.FormationConfig{
		FormationID: "my-bot",
		ServerURL:   os.Getenv("MUXI_SERVER_URL"), // proxy mode
		ClientKey:   os.Getenv("MUXI_CLIENT_KEY"),
		MaxRetries:  2,
	})

	chunks, errs := client.ChatStream(ctx, &muxi.ChatRequest{
		Message: "Tell me a short story about MUXI",
		UserID:  "user-123",
	})

	for {
		select {
		case chunk, ok := <-chunks:
			if !ok {
				return
			}
			switch chunk.Type {
			case "text":
				fmt.Print(chunk.Text)
			case "error":
				fmt.Printf("\n[error] %s\n", chunk.Error)
			case "done":
				fmt.Println("\n[done]")
			}
		case err, ok := <-errs:
			if ok && err != nil {
				fmt.Printf("\n[stream error] %v\n", err)
			}
			return
		}
	}
}
