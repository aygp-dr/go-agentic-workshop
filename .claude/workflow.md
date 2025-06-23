# Claude AI Agent Git Workflow Standards

This document defines how AI agents (Claude and others) should interact with git repositories to ensure consistent, traceable, and collaborative development.

## Workflow Setup

This workflow assumes:
1. **AI Agent Account**: A dedicated GitHub account (e.g., `aygp-dr`) for AI-assisted commits
2. **Human Account**: A separate account for human developer work (e.g., `jwalsh`)
3. **Co-authorship**: Claude is credited as co-author using `Co-authored-by` trailers

This separation provides:
- Clear isolation between human and AI contributions
- Proper attribution in git history
- Easy filtering and analysis of AI-generated code

## Quick Start

```bash
# Initialize git for AI agent workflows
git config --local commit.template .gitmessage.claude
git config --local notes.ref refs/notes/ai-agent
git config --local core.commentChar ";"

# Automatically detect and use current GitHub user
if command -v gh &> /dev/null; then
  GH_USER=$(gh api user --jq '.login')
  GH_EMAIL=$(gh api user --jq '.email // "\(.login)@users.noreply.github.com"')
  GH_NAME=$(gh api user --jq '.name // .login')
  
  git config --local user.name "$GH_NAME"
  git config --local user.email "$GH_EMAIL"
  
  echo "Configured git for: $GH_NAME <$GH_EMAIL>"
else
  echo "GitHub CLI not found. Set user manually:"
  echo "  git config --local user.name 'Your Name'"
  echo "  git config --local user.email 'your-github-username@users.noreply.github.com'"
fi
```

## Conventional Commits

All AI agents MUST use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types

- `feat`: New feature implementation
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, semicolons, etc.)
- `refactor`: Code refactoring without changing functionality
- `perf`: Performance improvements
- `test`: Adding or modifying tests
- `build`: Build system or dependency changes
- `ci`: CI/CD configuration changes
- `chore`: Other changes that don't modify src or test files
- `revert`: Reverts a previous commit

### AI-Specific Types

- `ai-assist`: AI-assisted code changes
- `ai-review`: AI-performed code review changes
- `ai-gen`: AI-generated code from specifications

## Required Git Trailers

Every commit by an AI agent MUST include these trailers:

```
Agent-ID: <agent-identifier>
Agent-Model: <model-name-version>
Agent-Task-ID: <task-identifier>
Agent-Confidence: <0.0-1.0>
Agent-Mode: <autonomous|assisted|review>
Human-Review-Required: <true|false>
```

### Optional Trailers

```
Co-authored-by: Claude (AI Assistant) <claude@anthropic.ai>
Agent-Context-Length: <number>
Agent-Temperature: <0.0-2.0>
Agent-Reasoning: <brief-explanation>
Performance-Impact: <positive|neutral|negative>
Security-Impact: <none|review-needed|improved>
Breaking-Change: <true|false>
Related-To: <issue-or-task-id>
```

## Commit Message Template

Create `.gitmessage.claude`:

```
# <type>[scope]: <subject> (max 50 chars)

# <body> (max 72 chars per line)
# Explain what and why, not how

# === AI AGENT METADATA (Required) ===
Agent-ID: claude-3-opus
Agent-Model: claude-3-opus-20240229
Agent-Task-ID: 
Agent-Confidence: 
Agent-Mode: 
Human-Review-Required: 

# === OPTIONAL TRAILERS ===
# Co-authored-by: Claude (AI Assistant) <claude@anthropic.ai>
# Agent-Context-Length: 
# Agent-Temperature: 
# Agent-Reasoning: 
# Performance-Impact: 
# Security-Impact: 
# Breaking-Change: 
# Related-To: 
```

## Git Notes Usage

### When to Use Notes

Use git notes for information that:
1. Is generated AFTER the commit
2. May change over time
3. Contains detailed analysis or logs
4. Includes review feedback

### Notes Namespaces

```bash
# Post-commit analysis
git notes --ref=ai/analysis add -m "..."

# Code review results
git notes --ref=ai/review add -m "..."

# Performance benchmarks
git notes --ref=ai/performance add -m "..."

# Security scan results
git notes --ref=ai/security add -m "..."

# Agent conversation logs
git notes --ref=ai/conversation add -m "..."
```

### Notes Format

Always use JSON for structured data:

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "agent_id": "claude-3-opus",
  "analysis_type": "security|performance|quality|review",
  "results": {
    "score": 0.95,
    "issues": [],
    "suggestions": []
  },
  "metadata": {}
}
```

## Complete Workflow Example

```bash
# 0. Switch to AI agent account
gh auth login --hostname github.com
# Select the AI agent account (e.g., aygp-dr)

# 1. AI agent makes changes
git add -A

# 2. Commit with conventional format and trailers
git commit -m "feat(auth): implement OAuth2 token refresh" -m "
Added automatic token refresh mechanism to prevent authentication
failures. The implementation uses a background timer that checks
token expiration and refreshes 5 minutes before expiry.

Agent-ID: claude-3-opus
Agent-Model: claude-3-opus-20240229
Agent-Task-ID: AUTH-2025-01-15-001
Agent-Confidence: 0.92
Agent-Mode: autonomous
Human-Review-Required: true
Agent-Reasoning: Implemented standard OAuth2 refresh pattern
Performance-Impact: neutral
Security-Impact: improved
Related-To: #1234
Co-authored-by: Claude (AI Assistant) <claude@anthropic.ai>"

# 3. Run automated analysis
./scripts/analyze-commit.sh HEAD

# 4. Add analysis results as note
git notes --ref=ai/analysis add -m '{
  "timestamp": "2025-01-15T10:35:00Z",
  "agent_id": "security-scanner-v2",
  "analysis_type": "security",
  "results": {
    "score": 1.0,
    "vulnerabilities": [],
    "suggestions": [
      "Consider implementing token rotation",
      "Add rate limiting to refresh endpoint"
    ]
  }
}'

# 5. Human review adds note
git notes --ref=ai/review add -m '{
  "timestamp": "2025-01-15T14:00:00Z",
  "reviewer": "alice@example.com",
  "status": "approved_with_suggestions",
  "comments": "Good implementation. Please add the suggested rate limiting."
}'
```

## Querying AI Commits

```bash
# Find all AI agent commits
git log --grep="^Agent-ID:"

# Find commits requiring human review
git log --grep="^Human-Review-Required: true"

# Find high-confidence AI commits
git log --grep="^Agent-Confidence: 0\.9"

# Show commits with notes
git log --show-notes=ai/analysis --show-notes=ai/review

# Export AI metrics
git log --format="%H %s%n%(trailers:key=Agent-ID,Agent-Confidence)" | grep -v "^$"

# Find commits co-authored by Claude
git log --grep="^Co-authored-by: Claude"

# Show commits by the AI agent account
git log --author="aygp-dr"
```

## Best Practices

### Account Management:
- ✅ Use a dedicated GitHub account for AI agent work (e.g., `aygp-dr`)
- ✅ Keep human work on separate account (e.g., `jwalsh`)
- ✅ Always include `Co-authored-by: Claude` trailer for attribution
- ✅ Use GitHub CLI (`gh auth login`) to switch between accounts

### DO:
- ✅ Always include ALL required trailers
- ✅ Use conventional commit format
- ✅ Add detailed reasoning in commit body
- ✅ Use notes for post-commit information
- ✅ Maintain consistent agent identities
- ✅ Include task/issue references

### DON'T:
- ❌ Modify commits after pushing (use notes instead)
- ❌ Skip human review when confidence < 0.95
- ❌ Include sensitive data in commits or notes
- ❌ Use generic commit messages
- ❌ Forget to sync notes with remotes

## Git Hooks

### Pre-commit Hook (`.git/hooks/pre-commit`)

```bash
#!/bin/bash
# Validate AI agent commits

if git config user.email | grep -q "@anthropic.ai"; then
  # Check for required trailers
  commit_msg=$(git log -1 --format=%B)
  
  required_trailers=(
    "Agent-ID:"
    "Agent-Model:"
    "Agent-Task-ID:"
    "Agent-Confidence:"
    "Agent-Mode:"
    "Human-Review-Required:"
  )
  
  for trailer in "${required_trailers[@]}"; do
    if ! echo "$commit_msg" | grep -q "^$trailer"; then
      echo "ERROR: Missing required trailer: $trailer"
      exit 1
    fi
  done
fi
```

## Integration with CI/CD

```yaml
# .github/workflows/ai-agent-checks.yml
name: AI Agent Commit Checks

on: [push, pull_request]

jobs:
  validate-ai-commits:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
          
      - name: Check AI commits
        run: |
          # Validate all commits with AI agent metadata
          git log --format="%H" --grep="^Agent-ID:" | while read hash; do
            ./scripts/validate-ai-commit.sh $hash
          done
```

## References

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Git Notes Documentation](https://git-scm.com/docs/git-notes)
- [Git Trailers](https://git-scm.com/docs/git-interpret-trailers)
- [GitHub CLI](https://cli.github.com/) - For account management
- [Co-authored-by commits](https://docs.github.com/en/pull-requests/committing-changes-to-your-project/creating-and-editing-commits/creating-a-commit-with-multiple-authors)

---

*This document is version 1.0.0 and should be updated as AI agent capabilities evolve.*
