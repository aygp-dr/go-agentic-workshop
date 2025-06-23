#!/bin/bash
# Switch between human and AI agent GitHub accounts

if [ "$1" = "ai" ] || [ "$1" = "agent" ]; then
    echo "🤖 Switching to AI agent account..."
    gh auth switch
    
    # Update git config
    GH_USER=$(gh api user --jq '.login')
    GH_EMAIL=$(gh api user --jq '.email // "\(.login)@users.noreply.github.com"')
    GH_NAME=$(gh api user --jq '.name // .login')
    
    git config --local user.name "$GH_NAME"
    git config --local user.email "$GH_EMAIL"
    
    echo "✅ Switched to: $GH_NAME <$GH_EMAIL>"
    
elif [ "$1" = "human" ]; then
    echo "👤 Switching to human account..."
    gh auth switch
    
    # Update git config
    GH_USER=$(gh api user --jq '.login')
    GH_EMAIL=$(gh api user --jq '.email // "\(.login)@users.noreply.github.com"')
    GH_NAME=$(gh api user --jq '.name // .login')
    
    git config --local user.name "$GH_NAME"
    git config --local user.email "$GH_EMAIL"
    
    echo "✅ Switched to: $GH_NAME <$GH_EMAIL>"
    
else
    echo "Usage: $0 [ai|agent|human]"
    echo ""
    echo "Current git config:"
    echo "  Name: $(git config user.name)"
    echo "  Email: $(git config user.email)"
    echo ""
    if command -v gh &> /dev/null; then
        echo "Current GitHub user:"
        gh api user --jq '"  Login: \(.login)\n  Name: \(.name // .login)"' 2>/dev/null || echo "  Not authenticated"
    fi
fi
