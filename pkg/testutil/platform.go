package testutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Platform represents the running environment
type Platform struct {
	OS            string
	Arch          string
	IsDocker      bool
	IsCodespace   bool
	IsReplit      bool
	IsCI          bool
	HasGPU        bool
	CloudProvider string
}

// GetPlatform detects the current platform
func GetPlatform() *Platform {
	p := &Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Detect container environment
	if _, err := os.Stat("/.dockerenv"); err == nil {
		p.IsDocker = true
	}

	// Detect GitHub Codespaces
	if os.Getenv("CODESPACES") == "true" {
		p.IsCodespace = true
		p.CloudProvider = "github"
	}

	// Detect Replit
	if os.Getenv("REPL_ID") != "" {
		p.IsReplit = true
		p.CloudProvider = "replit"
	}

	// Detect CI environment
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		p.IsCI = true
	}

	// Detect cloud provider
	if p.CloudProvider == "" {
		if isAWS() {
			p.CloudProvider = "aws"
		}
	}

	return p
}

// SkipIfPlatform skips test on specific platforms
func SkipIfPlatform(t *testing.T, platforms ...string) {
	t.Helper()
	current := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	for _, platform := range platforms {
		if strings.Contains(current, platform) {
			t.Skipf("Skipping test on platform: %s", current)
		}
	}
}

// RequirePlatform requires specific platform for test
func RequirePlatform(t *testing.T, platforms ...string) {
	t.Helper()
	current := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	for _, platform := range platforms {
		if strings.Contains(current, platform) {
			return
		}
	}

	t.Skipf("Test requires platform: %v, current: %s", platforms, current)
}

// PlatformSpecificValue returns different values based on platform
func PlatformSpecificValue(defaultVal interface{}, overrides map[string]interface{}) interface{} {
	key := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	// Check specific OS/Arch combination first
	if val, ok := overrides[key]; ok {
		return val
	}

	// Check OS only
	if val, ok := overrides[runtime.GOOS]; ok {
		return val
	}

	// Check arch only
	if val, ok := overrides[runtime.GOARCH]; ok {
		return val
	}

	return defaultVal
}

// GetTestTimeout returns platform-specific timeout
func GetTestTimeout() time.Duration {
	timeouts := map[string]interface{}{
		"arm64":        30 * time.Second, // ARM processors might be slower
		"freebsd":      20 * time.Second,
		"windows":      20 * time.Second,
		"darwin/arm64": 15 * time.Second, // M1/M2 Macs are fast
		"linux/amd64":  10 * time.Second, // Default fast timeout
	}

	duration := PlatformSpecificValue(10*time.Second, timeouts)
	return duration.(time.Duration)
}

// GetMemoryLimit returns platform-specific memory limit
func GetMemoryLimit() int64 {
	limits := map[string]interface{}{
		"arm64":     int64(2 << 30),   // 2GB for ARM
		"replit":    int64(512 << 20), // 512MB for Replit
		"codespace": int64(4 << 30),   // 4GB for Codespaces
	}

	p := GetPlatform()
	if p.IsReplit {
		return limits["replit"].(int64)
	}
	if p.IsCodespace {
		return limits["codespace"].(int64)
	}

	limit := PlatformSpecificValue(int64(4<<30), limits)
	return limit.(int64)
}

// isAWS detects if running on AWS
func isAWS() bool {
	// Check EC2 metadata service
	client := &http.Client{Timeout: 100 * time.Millisecond}
	resp, err := client.Get("http://169.254.169.254/latest/meta-data/")
	if err == nil {
		_ = resp.Body.Close()
		return true
	}
	return false
}

// TestConfig provides platform-specific test configuration
type TestConfig struct {
	UseLocalStack  bool
	UseOllama      bool
	AWSEndpoint    string
	PostgresURL    string
	MaxConcurrency int
	TestDataPath   string
}

// GetTestConfig returns platform-specific test configuration
func GetTestConfig() *TestConfig {
	p := GetPlatform()

	config := &TestConfig{
		UseLocalStack:  !p.IsCI,
		UseOllama:      p.OS != "windows" && !p.IsReplit,
		AWSEndpoint:    "http://localhost:4566",
		PostgresURL:    "postgres://workshop:workshop@localhost:5432/workshop?sslmode=disable",
		MaxConcurrency: 4,
		TestDataPath:   "./testdata",
	}

	// Platform-specific overrides
	if p.IsReplit {
		config.MaxConcurrency = 1 // Limited resources
		config.UseLocalStack = false
	}

	if p.IsCodespace {
		config.AWSEndpoint = "http://localstack:4566"
		config.PostgresURL = "postgres://workshop:workshop@postgres:5432/workshop?sslmode=disable"
	}

	if p.OS == "freebsd" {
		// FreeBSD might have different network config
		config.UseOllama = false // Ollama might not be available
	}

	if p.Arch == "arm64" {
		config.MaxConcurrency = 2 // Be conservative on ARM
	}

	// Environment variable overrides
	if endpoint := os.Getenv("AWS_ENDPOINT"); endpoint != "" {
		config.AWSEndpoint = endpoint
	}

	if pgURL := os.Getenv("DATABASE_URL"); pgURL != "" {
		config.PostgresURL = pgURL
	}

	return config
}

// SetupPlatformSpecificEnv sets up environment for tests
func SetupPlatformSpecificEnv(t *testing.T) func() {
	t.Helper()

	p := GetPlatform()
	originalEnv := make(map[string]string)

	// Save original environment
	envVars := []string{"GOMAXPROCS", "GOMEMLIMIT"}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Set platform-specific values
	if p.Arch == "arm64" {
		_ = os.Setenv("GOMAXPROCS", "2")
	}

	if p.IsReplit {
		_ = os.Setenv("GOMEMLIMIT", "512MiB")
	}

	// Return cleanup function
	return func() {
		for key, value := range originalEnv {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}
}

// DockerAvailable checks if Docker is available on the platform
func DockerAvailable() bool {
	p := GetPlatform()

	// Docker might not be available on some platforms
	if p.IsReplit || p.OS == "freebsd" {
		return false
	}

	// Try to contact Docker daemon
	client := &http.Client{Timeout: 1 * time.Second}

	// Try different Docker socket locations
	sockets := []string{
		"http://localhost:2375",             // Windows/Mac
		"http://unix:/var/run/docker.sock:", // Linux
	}

	for _, socket := range sockets {
		resp, err := client.Get(socket + "/_ping")
		if err == nil {
			_ = resp.Body.Close()
			return true
		}
	}

	return false
}

// RequireDocker skips test if Docker is not available
func RequireDocker(t *testing.T) {
	t.Helper()
	if !DockerAvailable() {
		t.Skip("Docker not available on this platform")
	}
}

// GetPlatformTestTags returns build tags for platform
func GetPlatformTestTags() []string {
	p := GetPlatform()
	tags := []string{"integration"}

	if p.OS == "linux" {
		tags = append(tags, "linux")
	}

	if p.Arch == "arm64" {
		tags = append(tags, "arm64")
	}

	if p.IsCI {
		tags = append(tags, "ci")
	}

	return tags
}

// WaitForService waits for a service to be ready with platform-specific timeout
func WaitForService(ctx context.Context, url string) error {
	timeout := GetTestTimeout()
	deadline := time.Now().Add(timeout)

	client := &http.Client{Timeout: 1 * time.Second}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := client.Get(url)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode < 500 {
					return nil
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	return fmt.Errorf("service at %s did not become ready within %v", url, timeout)
}
