#!/bin/bash
#
# Setup git hooks for Genus ORM
# Run this script after cloning the repository to install commit validation hooks

set -e

HOOKS_DIR=".git/hooks"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🔧 Setting up git hooks for Genus ORM..."

# Create commit-msg hook
cat > "$HOOKS_DIR/commit-msg" << 'EOF'
#!/bin/sh
#
# Git hook to prevent commits with AI tool mentions
# This hook checks the commit message for references to AI tools

commit_msg_file=$1
commit_msg=$(cat "$commit_msg_file")

# Check for AI tool mentions (case insensitive)
if echo "$commit_msg" | grep -iE "(🤖|Generated with|Co-Authored-By:)" > /dev/null; then
    echo "Error: Commit message contains references to AI tools."
    echo "Please remove the following patterns:"
    echo "  - 🤖 Generated with"
    echo "  - Co-Authored-By:"
    echo ""
    echo "Blocked commit message:"
    echo "---"
    cat "$commit_msg_file"
    echo "---"
    exit 1
fi

exit 0
EOF

# Make hook executable
chmod +x "$HOOKS_DIR/commit-msg"

echo "✅ Git hooks installed successfully!"
echo ""
echo "The following hooks are now active:"
echo "  - commit-msg: Validates commit messages (blocks AI tool mentions)"
echo ""
echo "To bypass a hook (not recommended): git commit --no-verify"
