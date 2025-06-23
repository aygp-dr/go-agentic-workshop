# Go Agentic Workshop - Building Production AI Agents

A hands-on workshop for GopherCon 2025 focused on building production-ready AI agents using Go, AWS Bedrock, and modern orchestration patterns.

## Prerequisites

- Go 1.23 or later installed
- AWS account with Bedrock access configured
- Docker and Docker Compose installed
- Basic knowledge of Go programming
- AWS CLI configured with appropriate credentials

## Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/aygp-dr/go-agentic-workshop.git
   cd go-agentic-workshop
   ```

2. Run the setup script to verify your environment:
   ```bash
   make setup
   ```

3. Start the required services:
   ```bash
   make docker-up
   ```

4. Run the agent:
   ```bash
   make run
   ```

## Workshop Content

For detailed workshop content, exercises, and architecture documentation, see [SETUP.org](SETUP.org).

## Project Structure

```
.
├── cmd/agent/          # Main agent application
├── pkg/                # Shared packages
│   ├── bedrock/       # AWS Bedrock client
│   ├── orchestrator/  # Agent orchestration logic
│   └── patterns/      # Common agent patterns
├── examples/          # Example implementations
├── benchmarks/        # Performance benchmarks
└── setup/            # Setup and configuration scripts
```

## Development

Run tests:
```bash
make test
```

Stop services:
```bash
make docker-down
```