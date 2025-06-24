# Go Agentic Workshop - Building Production AI Agents

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-draft-orange)

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

## Contributing

We welcome contributions! This project is developed and tested on multiple platforms.

### Supported Platforms

| Platform | Architecture | Go Version | Status |
|----------|--------------|------------|---------|
| Linux (Raspberry Pi) | ARM64/aarch64 | 1.23.0+ | ✅ Tested |
| Linux (x86_64) | AMD64 | 1.23.0+ | ✅ Supported |
| FreeBSD | AMD64 | 1.23.0+ | ✅ Supported |
| macOS | ARM64/AMD64 | 1.23.0+ | ✅ Supported |
| Windows | AMD64 | 1.23.0+ | 🔄 Experimental |

### Development Environment

The project is primarily developed on:
- **OS**: Linux 6.12.20+rpt-rpi-v8 (Debian-based)
- **Architecture**: ARM64/aarch64
- **Go Version**: 1.23.0
- **Platform**: Raspberry Pi

### Getting Started

1. Ensure you have Go 1.23 or later installed
2. Fork the repository
3. Create a feature branch (`git checkout -b feature/amazing-feature`)
4. Commit your changes (`git commit -m 'feat: add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions or modifications
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

## License

This project is licensed under the MIT License - see the LICENSE file for details.