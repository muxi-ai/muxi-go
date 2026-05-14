package main

import (
    "context"
    "fmt"
    "os"

    "github.com/muxi-ai/muxi-go"
)

// Deploys a formation bundle (non-streaming)
func main() {
    ctx := context.Background()

    server := muxi.NewServerClient(&muxi.ServerConfig{
        URL:       os.Getenv("MUXI_SERVER_URL"),
        KeyID:     os.Getenv("MUXI_KEY_ID"),
        SecretKey: os.Getenv("MUXI_SECRET_KEY"),
        MaxRetries: 3,
    })

    result, err := server.DeployFormation(ctx, &muxi.DeployRequest{
        FormationID: "my-bot",
        BundlePath:  "./my-bot.tar.gz",
        Version:     "1.0.0",
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Deployed %s on port %d (v%s)\n", result.ID, result.Port, result.Version)
}
