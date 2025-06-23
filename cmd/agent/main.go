package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aygp-dr/go-agentic-workshop/pkg/bedrock"
)

type Config struct {
	Region          string
	BedrockEndpoint string
	ModelID         string
}

func main() {
	fmt.Println("AI Agent Workshop Starting...")

	config := loadConfig()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down agent...")
		cancel()
	}()

	_, err := bedrock.NewClient(ctx, config.Region)
	if err != nil {
		log.Fatalf("Failed to create Bedrock client: %v", err)
	}

	fmt.Printf("Agent initialized with region: %s\n", config.Region)
	fmt.Printf("Using model: %s\n", config.ModelID)
	
	<-ctx.Done()
	fmt.Println("Agent stopped")
}

func loadConfig() Config {
	return Config{
		Region:          getEnv("AWS_REGION", "us-east-1"),
		BedrockEndpoint: getEnv("BEDROCK_ENDPOINT", ""),
		ModelID:         getEnv("MODEL_ID", "anthropic.claude-3-sonnet-20240229-v1:0"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}