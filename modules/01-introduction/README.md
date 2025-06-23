# Module 1: Introduction to AI Agents & Workflows

**Duration:** 1 hour  
**Difficulty:** Intermediate

## Learning Objectives

By the end of this module, you will:
- Understand the fundamental differences between RAG and Agentic patterns
- Identify when to use AI agents vs traditional approaches
- Design basic agent architectures
- Set up your development environment

## Pre-requisites Checklist

- [ ] Go 1.21+ installed
- [ ] Docker running
- [ ] AWS CLI configured
- [ ] Workshop repository cloned
- [ ] Environment validation passed

Run the validator:
```bash
go run cmd/validator/main.go
```

## Module Overview

### 1.1 Beyond RAG: Why Agents? (15 min)

**Key Concepts:**
- Limitations of static retrieval patterns
- Multi-step reasoning requirements
- Autonomous decision making
- State management in AI systems

### 1.2 Agent Architecture Components (20 min)

**Core Components:**
1. **LLM Engine** - The reasoning core
2. **Memory System** - Short and long-term memory
3. **Function Registry** - Available tools/actions
4. **State Manager** - Workflow orchestration

### 1.3 Hands-On: Your First Agent (20 min)

Build a simple tool-calling agent that can:
- Parse user intent
- Select appropriate tools
- Execute functions
- Return formatted results

### 1.4 Discussion & Q&A (5 min)

## Exercises

### Exercise 1.1: Environment Setup
**Time:** 5 minutes

1. Run the environment validator
2. Fix any missing dependencies
3. Start Docker services
4. Verify connectivity

### Exercise 1.2: Agent vs RAG Comparison
**Time:** 10 minutes

Compare outputs for the same query:
- RAG system: Simple retrieval + generation
- Agent system: Multi-step reasoning with tools

### Exercise 1.3: Build a Calculator Agent
**Time:** 15 minutes

Create an agent that can:
- Parse mathematical expressions
- Use calculator functions
- Show step-by-step reasoning

## Common Issues & Solutions

### Issue 1: Docker not running
```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker

# Windows
Start Docker Desktop from Start Menu
```

### Issue 2: AWS credentials missing
```bash
aws configure
# Enter your AWS Access Key ID
# Enter your AWS Secret Access Key
# Enter default region (us-east-1)
# Enter default output format (json)
```

### Issue 3: Go modules not downloading
```bash
go clean -modcache
go mod download
```

## Additional Resources

- [Agent Design Patterns](https://arxiv.org/abs/2309.07864)
- [ReAct: Reasoning and Acting](https://arxiv.org/abs/2210.03629)
- [Function Calling Best Practices](https://platform.openai.com/docs/guides/function-calling)

## Instructor Notes

**Time Management:**
- Keep introduction conceptual (no code)
- Focus on hands-on time
- Use pre-built examples if running behind

**Common Questions:**
1. "When should I use agents vs RAG?"
   - Use agents for: Multi-step tasks, tool usage, complex reasoning
   - Use RAG for: Simple Q&A, document search, static knowledge

2. "What about latency?"
   - Address in Module 4 (optimization techniques)
   - Mention caching, streaming, parallel execution

**Backup Plans:**
- If AWS is down: Use LocalStack
- If Docker fails: Use pre-built binaries
- If network issues: Use offline mode with cached responses