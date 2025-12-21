package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type ValidationResult struct {
	Component  string `json:"component"`
	Status     string `json:"status"`
	Version    string `json:"version,omitempty"`
	Message    string `json:"message,omitempty"`
	Required   bool   `json:"required"`
	FixCommand string `json:"fix_command,omitempty"`
}

type SystemInfo struct {
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	GoVersion   string `json:"go_version"`
	CPUs        int    `json:"cpus"`
	TotalMemory string `json:"total_memory"`
}

var results []ValidationResult

func main() {
	fmt.Println("🔍 Go Agentic Workshop - Environment Validator")
	fmt.Println("=" + strings.Repeat("=", 45))
	fmt.Println()

	ctx := context.Background()

	// System information
	sysInfo := getSystemInfo()
	fmt.Printf("📊 System: %s/%s, %d CPUs\n", sysInfo.OS, sysInfo.Arch, sysInfo.CPUs)
	fmt.Printf("💾 Go Version: %s\n", sysInfo.GoVersion)
	fmt.Println()

	// Core requirements
	fmt.Println("🔧 Checking Core Requirements:")
	validateGo()
	validateDocker()
	validateAWSCLI()
	validateMake()

	// Workshop specific tools
	fmt.Println("\n🛠️  Checking Workshop Tools:")
	validateOllama()
	validateLocalStack()
	validatePostgreSQL()

	// Network connectivity
	fmt.Println("\n🌐 Checking Network Connectivity:")
	validateNetwork()

	// AWS configuration
	fmt.Println("\n☁️  Checking AWS Configuration:")
	validateAWSConfig(ctx)

	// Disk space
	fmt.Println("\n💽 Checking Disk Space:")
	validateDiskSpace()

	// Generate report
	generateReport()
}

func getSystemInfo() SystemInfo {
	return SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		CPUs:      runtime.NumCPU(),
	}
}

func validateGo() {
	result := ValidationResult{
		Component: "Go",
		Required:  true,
	}

	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "❌"
		result.Message = "Go not found in PATH"
		result.FixCommand = getGoInstallCommand()
	} else {
		version := strings.TrimSpace(string(output))
		result.Status = "✅"
		result.Version = extractVersion(version, "go")

		// Check minimum version
		if !isVersionSupported(result.Version, "1.21") {
			result.Status = "⚠️"
			result.Message = "Go version 1.21+ required"
			result.FixCommand = getGoInstallCommand()
		}
	}

	results = append(results, result)
	printResult(result)
}

func validateDocker() {
	result := ValidationResult{
		Component: "Docker",
		Required:  true,
	}

	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "❌"
		result.Message = "Docker not found"
		result.FixCommand = getDockerInstallCommand()
	} else {
		// Check if Docker daemon is running
		cmd = exec.Command("docker", "ps")
		if err := cmd.Run(); err != nil {
			result.Status = "⚠️"
			result.Message = "Docker daemon not running"
			result.FixCommand = "sudo systemctl start docker || open -a Docker"
		} else {
			result.Status = "✅"
			result.Version = extractVersion(string(output), "Docker version")
		}
	}

	results = append(results, result)
	printResult(result)
}

func validateAWSCLI() {
	result := ValidationResult{
		Component: "AWS CLI",
		Required:  true,
	}

	cmd := exec.Command("aws", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "❌"
		result.Message = "AWS CLI not found"
		result.FixCommand = "curl \"https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip\" -o \"awscliv2.zip\" && unzip awscliv2.zip && sudo ./aws/install"
	} else {
		result.Status = "✅"
		result.Version = extractVersion(string(output), "aws-cli/")
	}

	results = append(results, result)
	printResult(result)
}

func validateMake() {
	result := ValidationResult{
		Component: "Make",
		Required:  true,
	}

	cmd := exec.Command("make", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "❌"
		result.Message = "Make not found"
		result.FixCommand = getMakeInstallCommand()
	} else {
		result.Status = "✅"
		result.Version = strings.Split(string(output), "\n")[0]
	}

	results = append(results, result)
	printResult(result)
}

func validateOllama() {
	result := ValidationResult{
		Component: "Ollama",
		Required:  false,
	}

	// Check if Ollama is installed
	cmd := exec.Command("ollama", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "⚠️"
		result.Message = "Ollama not installed (optional for local LLM testing)"
		result.FixCommand = "curl -fsSL https://ollama.ai/install.sh | sh"
	} else {
		result.Status = "✅"
		result.Version = strings.TrimSpace(string(output))

		// Check if Ollama is running
		resp, err := http.Get("http://localhost:11434/api/tags")
		if err != nil {
			result.Status = "⚠️"
			result.Message = "Ollama installed but not running"
			result.FixCommand = "ollama serve"
		} else {
			_ = resp.Body.Close()
		}
	}

	results = append(results, result)
	printResult(result)
}

func validateLocalStack() {
	result := ValidationResult{
		Component: "LocalStack",
		Required:  false,
	}

	// Check if LocalStack container exists
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "localstack") {
		result.Status = "✅"
		result.Version = "Docker container"
	} else {
		result.Status = "ℹ️"
		result.Message = "LocalStack container not found (will be created by docker-compose)"
	}

	results = append(results, result)
	printResult(result)
}

func validatePostgreSQL() {
	result := ValidationResult{
		Component: "PostgreSQL (pgvector)",
		Required:  false,
	}

	// Check for pgvector container
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "pgvector") {
		result.Status = "✅"
		result.Version = "Docker image available"
	} else {
		result.Status = "ℹ️"
		result.Message = "pgvector image not found (will be pulled by docker-compose)"
	}

	results = append(results, result)
	printResult(result)
}

func validateNetwork() {
	endpoints := map[string]string{
		"GitHub":      "https://api.github.com",
		"AWS":         "https://aws.amazon.com",
		"Docker Hub":  "https://hub.docker.com",
		"Go Packages": "https://proxy.golang.org",
	}

	for name, url := range endpoints {
		result := ValidationResult{
			Component: fmt.Sprintf("Network: %s", name),
			Required:  true,
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			result.Status = "❌"
			result.Message = fmt.Sprintf("Cannot reach %s", url)
		} else {
			_ = resp.Body.Close()
			result.Status = "✅"
		}

		results = append(results, result)
		printResult(result)
	}
}

func validateAWSConfig(ctx context.Context) {
	result := ValidationResult{
		Component: "AWS Credentials",
		Required:  true,
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		result.Status = "❌"
		result.Message = "AWS credentials not configured"
		result.FixCommand = "aws configure"
	} else {
		// Try to create SQS client as a test
		client := sqs.NewFromConfig(cfg)
		_, err = client.ListQueues(ctx, &sqs.ListQueuesInput{})
		if err != nil && strings.Contains(err.Error(), "credentials") {
			result.Status = "⚠️"
			result.Message = "AWS credentials found but may be invalid"
		} else {
			result.Status = "✅"
			result.Version = cfg.Region
		}
	}

	results = append(results, result)
	printResult(result)
}

func validateDiskSpace() {
	result := ValidationResult{
		Component: "Disk Space",
		Required:  true,
	}

	// This is a simplified check - in production you'd use syscall
	result.Status = "ℹ️"
	result.Message = "Ensure at least 10GB free space for workshop materials"

	results = append(results, result)
	printResult(result)
}

func printResult(result ValidationResult) {
	status := result.Status
	component := result.Component

	if result.Required {
		component += " (required)"
	}

	fmt.Printf("%s %-30s", status, component)

	if result.Version != "" {
		fmt.Printf(" %s", result.Version)
	}

	if result.Message != "" {
		fmt.Printf("\n   └─ %s", result.Message)
	}

	if result.FixCommand != "" && result.Status != "✅" {
		fmt.Printf("\n   └─ Fix: %s", result.FixCommand)
	}

	fmt.Println()
}

func generateReport() {
	fmt.Println("\n" + strings.Repeat("=", 50))

	required := 0
	requiredOK := 0
	optional := 0
	optionalOK := 0

	for _, r := range results {
		if r.Required {
			required++
			if r.Status == "✅" {
				requiredOK++
			}
		} else {
			optional++
			if r.Status == "✅" {
				optionalOK++
			}
		}
	}

	fmt.Printf("\n📋 Summary: %d/%d required components OK\n", requiredOK, required)
	fmt.Printf("           %d/%d optional components OK\n", optionalOK, optional)

	if requiredOK < required {
		fmt.Println("\n⚠️  Some required components are missing!")
		fmt.Println("Please install missing components before starting the workshop.")
	} else {
		fmt.Println("\n✅ Your environment is ready for the workshop!")
	}

	// Save detailed report
	reportFile := "validation-report.json"
	data, _ := json.MarshalIndent(map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"system":    getSystemInfo(),
		"results":   results,
	}, "", "  ")

	if err := os.WriteFile(reportFile, data, 0644); err == nil {
		fmt.Printf("\n📄 Detailed report saved to: %s\n", reportFile)
	}
}

// Helper functions
func extractVersion(output, prefix string) string {
	output = strings.TrimSpace(output)
	if strings.Contains(output, prefix) {
		parts := strings.Split(output, " ")
		for i, part := range parts {
			if strings.Contains(part, prefix) && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return output
}

func isVersionSupported(version, minimum string) bool {
	// Simplified version comparison
	return version >= minimum
}

func getGoInstallCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install go || visit https://go.dev/dl/"
	case "linux":
		return "sudo snap install go --classic || visit https://go.dev/dl/"
	case "windows":
		return "choco install golang || visit https://go.dev/dl/"
	default:
		return "visit https://go.dev/dl/"
	}
}

func getDockerInstallCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install --cask docker || visit https://docs.docker.com/desktop/mac/install/"
	case "linux":
		return "curl -fsSL https://get.docker.com | sh"
	case "windows":
		return "visit https://docs.docker.com/desktop/windows/install/"
	default:
		return "visit https://docs.docker.com/get-docker/"
	}
}

func getMakeInstallCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "xcode-select --install || brew install make"
	case "linux":
		return "sudo apt-get install make || sudo yum install make"
	case "windows":
		return "choco install make || visit https://gnuwin32.sourceforge.net/packages/make.htm"
	default:
		return "install make for your platform"
	}
}
