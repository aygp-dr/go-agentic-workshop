# Benchmark Suite

This directory contains benchmark tests for the agent platform. The benchmarks are designed to measure performance metrics such as latency, throughput, and cost efficiency.

## Running Benchmarks

### Basic Usage

To run all benchmarks:

```bash
go test -bench=. ./benchmarks
```

To run a specific benchmark:

```bash
go test -bench=BenchmarkLLMLatency ./benchmarks
```

### Advanced Options

Benchmarks can be run with various flags to control their behavior:

- `-benchtime=5s`: Run each benchmark for 5 seconds (default is 1s)
- `-benchmem`: Include memory allocation statistics
- `-count=5`: Run each benchmark 5 times to get more stable results
- `-cpu=1,2,4`: Run benchmarks with different GOMAXPROCS values

Example:

```bash
go test -bench=. -benchmem -count=3 -timeout=30m ./benchmarks
```

### Platform-Specific Benchmarks

The benchmark suite includes platform-specific tests that adjust parameters based on the detected operating system and architecture. These benchmarks are particularly useful for testing performance across different environments:

```bash
go test -bench=BenchmarkPlatformSpecific ./benchmarks
```

## Interpreting Results

Benchmark results include several metrics:

- **operations/sec**: Number of operations completed per second
- **ns/op**: Average time per operation in nanoseconds
- **B/op**: Average bytes allocated per operation
- **allocs/op**: Average number of allocations per operation

### Custom Metrics

This benchmark suite also reports several custom metrics:

- **tokens/op**: Number of tokens processed per operation
- **tokens/sec**: Token throughput per second
- **$/op**: Estimated cost per operation
- **cache_hit_%**: Percentage of cache hits (for caching benchmarks)
- **concurrency**: Concurrency level used for the benchmark
- **steps/op**: Number of workflow steps per operation

### Example Output

```
BenchmarkLLMLatency/Small-8                1000      1003249 ns/op      25 tokens/op
BenchmarkLLMLatency/Medium-8                500      3008362 ns/op     250 tokens/op
BenchmarkLLMLatency/Large-8                 100     10123456 ns/op    1000 tokens/op
BenchmarkLLMThroughput/Parallel_4_Small-8  5000       203611 ns/op   1968.7 tokens/sec
```

## Available Benchmarks

1. **BenchmarkLLMLatency**: Measures response time across different prompt sizes
2. **BenchmarkLLMThroughput**: Measures tokens per second processing rate with different parallelism levels
3. **BenchmarkFunctionCalling**: Measures function execution overhead for different function types
4. **BenchmarkWorkflowExecution**: Measures complete workflow performance with varying complexity
5. **BenchmarkMemoryUsage**: Measures memory consumption patterns with different state sizes
6. **BenchmarkPlatformSpecific**: Runs platform-aware benchmarks with optimizations for different OS/arch combinations
7. **BenchmarkCostOptimization**: Measures cost-performance tradeoffs with different caching strategies

## Platform Support

The benchmarks are designed to work across multiple platforms including:

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)
- FreeBSD (amd64)

Each platform may have different performance characteristics, and the platform-specific benchmarks adjust accordingly.

## Extending Benchmarks

To add a new benchmark:

1. Create a new function with the `Benchmark` prefix
2. Use the testing.B parameter to control the benchmark
3. Add proper setup and teardown code
4. Use b.ResetTimer() before the actual benchmark code
5. Use custom reporting metrics as needed

Example:

```go
func BenchmarkExample(b *testing.B) {
    // Setup code
    client := NewClient()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Benchmark code
        client.DoSomething()
    }
    
    // Report custom metrics
    b.ReportMetric(float64(customMetric), "custom_metric/op")
}
```