package bedrock

import (
    "context"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// Client wraps AWS Bedrock Runtime client
type Client struct {
    bedrock *bedrockruntime.Client
    config  *Config
}

// Config holds Bedrock configuration
type Config struct {
    Region      string
    ModelID     string
    MaxTokens   int
    Temperature float32
}

// NewClient creates a new Bedrock client
func NewClient(ctx context.Context, region string) (*Client, error) {
    awsCfg, err := config.LoadDefaultConfig(ctx, 
        config.WithRegion(region))
    if err != nil {
        return nil, err
    }
    
    return &Client{
        bedrock: bedrockruntime.NewFromConfig(awsCfg),
        config:  &Config{Region: region},
    }, nil
}
