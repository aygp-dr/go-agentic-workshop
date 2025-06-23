#!/bin/bash
# scripts/claude-init.sh - Initialize git repository for AI agent workflows

set -e

echo "🤖 Initializing git repository for Claude AI agent workflows..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository. Run 'git init' first."
    exit 1
fi

# Set up git user from GitHub CLI
echo "📝 Setting up git user identity..."
if command -v gh &> /dev/null; then
    GH_USER=$(gh api user --jq '.login')
    GH_EMAIL=$(gh api user --jq '.email // "\(.login)@users.noreply.github.com"')
    GH_NAME=$(gh api user --jq '.name // .login')
    
    git config --local user.name "$GH_NAME"
    git config --local user.email "$GH_EMAIL"
    
    echo "✅ Configured git for: $GH_NAME <$GH_EMAIL>"
else
    echo "⚠️  GitHub CLI not found. Using current git config:"
    echo "   Name: $(git config user.name)"
    echo "   Email: $(git config user.email)"
    echo ""
    echo "   To set manually:"
    echo "   git config --local user.name 'Your Name'"
    echo "   git config --local user.email 'your-github-username@users.noreply.github.com'"
fi

# Configure git settings
echo "⚙️  Configuring git settings..."
git config --local commit.template .gitmessage.claude
git config --local notes.ref refs/notes/ai-agent
git config --local core.commentChar ";"

# Create commit message template
echo "📄 Creating commit message template..."
cat > .gitmessage.claude << 'TEMPLATE'
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
TEMPLATE

# Create validation script
echo "📜 Creating validation script..."
cat > scripts/validate-ai-commit.sh << 'VALIDATOR'
#!/bin/bash
# Validate AI agent commit

commit_hash=$1
commit_msg=$(git log -1 --format=%B $commit_hash)

# Check if this is an AI commit by looking for Agent-ID
if echo "$commit_msg" | grep -q "^Agent-ID:"; then
  required_trailers=(
    "Agent-ID:"
    "Agent-Model:"
    "Agent-Task-ID:"
    "Agent-Confidence:"
    "Agent-Mode:"
    "Human-Review-Required:"
  )
  
  errors=0
  for trailer in "${required_trailers[@]}"; do
    if ! echo "$commit_msg" | grep -q "^$trailer"; then
      echo "❌ Missing required trailer: $trailer"
      ((errors++))
    fi
  done
  
  if [ $errors -gt 0 ]; then
    echo "❌ Commit $commit_hash failed validation"
    exit 1
  else
    echo "✅ Commit $commit_hash passed validation"
  fi
else
  echo "ℹ️  Commit $commit_hash is not an AI agent commit"
fi
VALIDATOR
chmod +x scripts/validate-ai-commit.sh

# Create analysis script placeholder
echo "📊 Creating analysis script placeholder..."
cat > scripts/analyze-commit.sh << 'ANALYZER'
#!/bin/bash
# Analyze commit and add notes

commit_hash=${1:-HEAD}
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Add your analysis logic here
analysis_result='{
  "timestamp": "'$timestamp'",
  "agent_id": "code-analyzer-v1",
  "analysis_type": "quality",
  "results": {
    "score": 0.95,
    "issues": [],
    "suggestions": []
  }
}'

# Add analysis as a note
git notes --ref=ai/analysis add -f -m "$analysis_result" $commit_hash
echo "✅ Analysis added to commit $commit_hash"
ANALYZER
chmod +x scripts/analyze-commit.sh

# Set up git aliases
echo "🔧 Setting up helpful git aliases..."
git config --local alias.ai-log "log --show-notes=ai/analysis --show-notes=ai/review --grep='^Agent-ID:'"
git config --local alias.ai-pending "log --grep='^Human-Review-Required: true' --grep='^Agent-ID:' --all-match"
git config --local alias.ai-stats "shortlog -sn --grep='^Agent-ID:'"

echo ""
echo "✅ Git repository initialized for AI agent workflows!"
echo ""
echo "📋 Next steps:"
echo "  1. Review .gitmessage.claude and customize if needed"
echo "  2. Test with: git commit --allow-empty"
echo "  3. View AI commits: git ai-log"
echo "  4. Find pending reviews: git ai-pending"
echo "  5. See AI commit stats: git ai-stats"
echo "  6. Switch accounts: ./scripts/switch-account.sh [ai|human]"
echo ""
echo "📚 Full documentation available in .claude/workflow.md"
