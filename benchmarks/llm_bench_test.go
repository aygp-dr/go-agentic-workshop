package benchmarks

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aygp-dr/go-agentic-workshop/pkg/testutil"
)

// BenchmarkLLMLatency measures response time across different prompt sizes
func BenchmarkLLMLatency(b *testing.B) {
	prompts := map[string]string{
		"Small":  generatePrompt(100),   // ~25 tokens
		"Medium": generatePrompt(1000),  // ~250 tokens
		"Large":  generatePrompt(4000),  // ~1000 tokens
		"XLarge": generatePrompt(16000), // ~4000 tokens
	}

	client := testutil.NewMockLLMClient()
	ctx := context.Background()

	for name, prompt := range prompts {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Call(ctx, prompt)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(len(prompt))/4, "tokens/op")
		})
	}
}

// BenchmarkLLMThroughput measures tokens per second processing rate
func BenchmarkLLMThroughput(b *testing.B) {
	scenarios := []struct {
		name       string
		promptSize int
		parallel   int
	}{
		{"Sequential_Small", 100, 1},
		{"Sequential_Large", 4000, 1},
		{"Parallel_2_Small", 100, 2},
		{"Parallel_4_Small", 100, 4},
		{"Parallel_2_Large", 4000, 2},
		{"Parallel_4_Large", 4000, 4},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			client := testutil.NewMockLLMClient()
			client.SetLatency(50 * time.Millisecond) // Simulate realistic latency

			prompt := generatePrompt(scenario.promptSize)
			ctx := context.Background()

			b.SetParallelism(scenario.parallel)
			b.ResetTimer()

			startTime := time.Now()
			totalTokens := 0

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					response, err := client.Call(ctx, prompt)
					if err != nil {
						b.Fatal(err)
					}
					totalTokens += (len(prompt) + len(response)) / 4
				}
			})

			duration := time.Since(startTime)
			tokensPerSecond := float64(totalTokens) / duration.Seconds()
			b.ReportMetric(tokensPerSecond, "tokens/sec")
		})
	}
}

// BenchmarkFunctionCalling measures function execution overhead
func BenchmarkFunctionCalling(b *testing.B) {
	registry := testutil.NewMockFunctionRegistry()

	// Register test functions with different complexities
	registry.Register("simple", func(args map[string]interface{}) (interface{}, error) {
		return args["input"], nil
	})

	registry.Register("compute", func(args map[string]interface{}) (interface{}, error) {
		// Simulate computation
		sum := 0.0
		for i := 0; i < 1000; i++ {
			sum += float64(i)
		}
		return sum, nil
	})

	registry.Register("io_bound", func(args map[string]interface{}) (interface{}, error) {
		// Simulate I/O operation
		time.Sleep(1 * time.Millisecond)
		return "done", nil
	})

	functions := []string{"simple", "compute", "io_bound"}
	ctx := context.Background()

	for _, fn := range functions {
		b.Run(fn, func(b *testing.B) {
			args := map[string]interface{}{"input": "test"}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := registry.Execute(ctx, fn, args)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkWorkflowExecution measures complete workflow performance
func BenchmarkWorkflowExecution(b *testing.B) {
	workflows := []struct {
		name  string
		steps int
	}{
		{"Simple_3_Steps", 3},
		{"Medium_10_Steps", 10},
		{"Complex_25_Steps", 25},
	}

	for _, workflow := range workflows {
		b.Run(workflow.name, func(b *testing.B) {
			client := testutil.NewMockLLMClient()
			registry := testutil.NewMockFunctionRegistry()
			state := testutil.NewMockWorkflowState()

			// Setup mock functions
			for i := 0; i < workflow.steps; i++ {
				fnName := fmt.Sprintf("step_%d", i)
				registry.Register(fnName, func(args map[string]interface{}) (interface{}, error) {
					return fmt.Sprintf("result_%v", args["step"]), nil
				})
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ctx := context.Background()

				// Simulate workflow execution
				for step := 0; step < workflow.steps; step++ {
					// LLM decides next action
					prompt := fmt.Sprintf("Execute step %d of workflow", step)
					response, err := client.Call(ctx, prompt)
					if err != nil {
						b.Fatal(err)
					}

					// Execute function
					fnName := fmt.Sprintf("step_%d", step)
					result, err := registry.Execute(ctx, fnName, map[string]interface{}{
						"step": step,
					})
					if err != nil {
						b.Fatal(err)
					}

					// Update state
					state.Set(fnName, result)
					state.Set("last_response", response)
				}
			}

			b.ReportMetric(float64(workflow.steps), "steps/op")
			b.ReportMetric(client.GetEstimatedCost()*float64(b.N), "$/total")
		})
	}
}

// BenchmarkMemoryUsage measures memory consumption patterns
func BenchmarkMemoryUsage(b *testing.B) {
	scenarios := []struct {
		name         string
		historySize  int
		stateEntries int
	}{
		{"Small_History", 10, 5},
		{"Medium_History", 100, 20},
		{"Large_History", 1000, 50},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				state := testutil.NewMockWorkflowState()

				// Simulate workflow with history
				for j := 0; j < scenario.historySize; j++ {
					for k := 0; k < scenario.stateEntries; k++ {
						key := fmt.Sprintf("state_%d_%d", j, k)
						value := fmt.Sprintf("value_%d_%d_with_some_content", j, k)
						state.Set(key, value)
					}
				}

				// Force JSON serialization to measure full memory impact
				_, err := state.ToJSON()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkPlatformSpecific runs platform-aware benchmarks
func BenchmarkPlatformSpecific(b *testing.B) {
	p := testutil.GetPlatform()
	config := testutil.GetTestConfig()

	// Skip test on unsupported platforms - this is just an example
	// In real benchmarks, we should support all platforms with appropriate settings
	unsupportedPlatforms := []string{}
	for _, platform := range unsupportedPlatforms {
		if strings.Contains(p.OS, platform) {
			b.Skipf("Skipping on unsupported platform: %s", p.OS)
		}
	}

	// Multiple benchmark configurations for different platforms
	benchConfigs := []struct {
		name        string
		promptSize  int
		concurrency int
		platformKey string
	}{
		{
			name:        "Small_Prompt_Standard",
			promptSize:  500,
			concurrency: config.MaxConcurrency,
			platformKey: "", // All platforms
		},
		{
			name:        "Large_Prompt_Standard",
			promptSize:  2000,
			concurrency: config.MaxConcurrency,
			platformKey: "", // All platforms
		},
		{
			name:        "FreeBSD_Optimized",
			promptSize:  1000,
			concurrency: 2, // Lower concurrency for FreeBSD
			platformKey: "freebsd",
		},
		{
			name:        "Darwin_Optimized",
			promptSize:  4000,
			concurrency: 8, // Higher concurrency for macOS
			platformKey: "darwin",
		},
	}

	for _, cfg := range benchConfigs {
		// Skip benchmarks not meant for this platform
		if cfg.platformKey != "" && !strings.Contains(p.OS, cfg.platformKey) {
			continue
		}

		b.Run(cfg.name, func(b *testing.B) {
			// Set appropriate concurrency
			concurrency := cfg.concurrency
			if concurrency <= 0 {
				concurrency = config.MaxConcurrency
			}
			b.SetParallelism(concurrency)

			client := testutil.NewMockLLMClient()

			// Platform-specific latency settings
			latencies := map[string]time.Duration{
				"darwin/arm64":  20 * time.Millisecond,   // M1/M2 fast
				"darwin/amd64": 30 * time.Millisecond,    // Intel Mac
				"linux/amd64":  50 * time.Millisecond,    // Standard Linux
				"linux/arm64":  100 * time.Millisecond,   // RPi slower
				"freebsd/amd64": 40 * time.Millisecond,   // FreeBSD performance
				"windows/amd64": 60 * time.Millisecond,   // Windows performance
			}

			// Apply platform-specific latency or use default
			platformKey := fmt.Sprintf("%s/%s", p.OS, p.Arch)
			if latency, ok := latencies[platformKey]; ok {
				client.SetLatency(latency)
			} else {
				// Default latency if platform not specifically configured
				client.SetLatency(70 * time.Millisecond)
			}

			b.ResetTimer()

			ctx := context.Background()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					prompt := generatePrompt(cfg.promptSize)
					response, err := client.Call(ctx, prompt)
					if err != nil {
						b.Fatal(err)
					}
					
					// Validate we got a meaningful response
					if len(response) == 0 {
						b.Fatal("Empty response received")
					}
				}
			})

			// Report platform-specific metrics
			b.ReportMetric(float64(concurrency), "concurrency")
			b.ReportMetric(float64(cfg.promptSize)/4, "prompt_tokens")
			b.ReportMetric(client.GetEstimatedCost()/float64(b.N), "$/op")
			// Simulated throughput calculation based on platform performance
			processingSpeed := float64(1000) / latencies[platformKey].Seconds()
			b.ReportMetric(processingSpeed, "tokens/sec")
		})
	}
}

// Helper function to generate prompts of specific sizes
func generatePrompt(chars int) string {
	template := "Please analyze the following text and provide insights: "
	padding := chars - len(template)
	if padding <= 0 {
		return template
	}

	// Generate realistic text-like content
	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog"}
	result := template

	for len(result) < chars {
		result += words[len(result)%len(words)] + " "
	}

	return result[:chars]
}

// BenchmarkCostOptimization measures cost-performance tradeoffs
func BenchmarkCostOptimization(b *testing.B) {
	strategies := []struct {
		name         string
		cacheHitRate float64
		batchSize    int
		modelSize    string
	}{
		{"NoCache_Single_Large", 0.0, 1, "large"},
		{"Cache50_Single_Large", 0.5, 1, "large"},
		{"Cache90_Single_Large", 0.9, 1, "large"},
		{"NoCache_Batch10_Large", 0.0, 10, "large"},
		{"Cache50_Batch10_Small", 0.5, 10, "small"},
	}

	modelCosts := map[string]float64{
		"small": 0.00001, // $0.01 per 1K tokens
		"large": 0.00006, // $0.06 per 1K tokens
	}

	for _, strategy := range strategies {
		b.Run(strategy.name, func(b *testing.B) {
			client := testutil.NewMockLLMClient()
			client.SetCostPerToken(modelCosts[strategy.modelSize])
			
			cache := make(map[string]string)
			cacheHits := 0
			totalCost := 0.0

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				prompts := make([]string, strategy.batchSize)
				for j := 0; j < strategy.batchSize; j++ {
					// Some prompts repeat to simulate cache scenarios
					divisor := strategy.batchSize/2
					if divisor <= 0 {
						divisor = 1
					}
					prompts[j] = fmt.Sprintf("prompt_%d", j%divisor)
				}

				for _, prompt := range prompts {
					// Check cache
					if cached, ok := cache[prompt]; ok && Random() < strategy.cacheHitRate {
						cacheHits++
						_ = cached // Use cached response
						continue
					}

					// Call LLM
					ctx := context.Background()
					response, err := client.Call(ctx, prompt)
					if err != nil {
						b.Fatal(err)
					}

					// Update cache
					cache[prompt] = response
				}
			}

			totalCost = client.GetEstimatedCost()
			b.ReportMetric(totalCost/float64(b.N), "$/op")
			b.ReportMetric(float64(cacheHits)/float64(b.N*strategy.batchSize)*100, "cache_hit_%")
		})
	}
}
