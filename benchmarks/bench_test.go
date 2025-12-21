package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/aygp-dr/go-agentic-workshop/pkg/testutil"
)

// BenchmarkLLMRequest benchmarks the performance of LLM requests with different model sizes.
// It measures the time taken to complete a request and reports tokens processed per operation.
func BenchmarkLLMRequest(b *testing.B) {
	models := map[string]struct {
		tokenCount    int
		responseRatio float64
		latency       time.Duration
	}{
		"Small Model": {
			tokenCount:    100,
			responseRatio: 1.5,
			latency:       20 * time.Millisecond,
		},
		"Large Model": {
			tokenCount:    1000,
			responseRatio: 2.0,
			latency:       100 * time.Millisecond,
		},
	}

	for name, model := range models {
		b.Run(name, func(b *testing.B) {
			client := testutil.NewMockLLMClient()
			client.SetLatency(model.latency)

			// Create a prompt with the specified token count (approximated as 4 chars per token)
			prompt := generatePrompt(model.tokenCount * 4)

			ctx := context.Background()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				response, err := client.Call(ctx, prompt)
				if err != nil {
					b.Fatal(err)
				}

				// Make sure we actually got a response
				if len(response) == 0 {
					b.Fatal("Empty response received from LLM client")
				}
			}

			// Report custom metrics
			promptTokens := len(prompt) / 4
			b.ReportMetric(float64(promptTokens), "prompt_tokens/op")
			b.ReportMetric(client.GetEstimatedCost()/float64(b.N), "$/op")
		})
	}
}

// BenchmarkPlatformAware runs benchmarks that adapt to the current platform
func BenchmarkPlatformAware(b *testing.B) {
	p := testutil.GetPlatform()
	config := testutil.GetTestConfig()

	// Set concurrency based on platform
	concurrency := config.MaxConcurrency

	b.Run("PlatformOptimized", func(b *testing.B) {
		client := testutil.NewMockLLMClient()

		// Platform-specific latencies
		latencies := map[string]time.Duration{
			"darwin/arm64":  20 * time.Millisecond, // M1/M2 Macs are fast
			"linux/amd64":   50 * time.Millisecond, // Standard performance
			"linux/arm64":   80 * time.Millisecond, // Slower ARM devices
			"freebsd/amd64": 40 * time.Millisecond, // FreeBSD specific setting
			"windows/amd64": 60 * time.Millisecond, // Windows performance
		}

		platformKey := p.OS + "/" + p.Arch
		if latency, ok := latencies[platformKey]; ok {
			client.SetLatency(latency)
		} else {
			// Default latency if platform not specifically configured
			client.SetLatency(70 * time.Millisecond)
		}

		b.SetParallelism(concurrency)
		ctx := context.Background()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				prompt := generatePrompt(500) // ~125 tokens
				response, err := client.Call(ctx, prompt)
				if err != nil {
					b.Fatal(err)
				}

				if len(response) == 0 {
					b.Fatal("Empty response received")
				}
			}
		})

		// Report platform-specific metrics
		b.ReportMetric(float64(concurrency), "concurrency")
		b.ReportMetric(client.GetEstimatedCost()/float64(b.N), "$/op")
	})
}
