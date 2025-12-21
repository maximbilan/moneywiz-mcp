#!/bin/bash

# MoneyWiz MCP Server Setup Script
# This script helps configure Claude Desktop to use the MoneyWiz MCP server

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}MoneyWiz MCP Server Setup${NC}"
echo "================================"
echo ""

# Get absolute paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_PATH="$SCRIPT_DIR/moneywiz-mcp"
DB_PATH="$SCRIPT_DIR/iMoneyWiz-Data-Backup-2025_12_21-17_23"
CONFIG_FILE="$HOME/Library/Application Support/Claude/claude_desktop_config.json"

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}Error: Binary not found at $BINARY_PATH${NC}"
    echo "Please build the server first: go build -o moneywiz-mcp ./cmd/main.go"
    exit 1
fi

# Make binary executable
chmod +x "$BINARY_PATH"
echo -e "${GREEN}✓${NC} Binary is executable"

# Check if database folder exists
if [ ! -d "$DB_PATH" ]; then
    echo -e "${YELLOW}Warning: Database folder not found at $DB_PATH${NC}"
    echo "You may need to update the path in the configuration file."
else
    echo -e "${GREEN}✓${NC} Database folder found"
fi

# Create config directory if it doesn't exist
mkdir -p "$HOME/Library/Application Support/Claude"
echo -e "${GREEN}✓${NC} Config directory ready"

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Creating new configuration file..."
    cat > "$CONFIG_FILE" << EOF
{
  "mcpServers": {
    "moneywiz": {
      "command": "$BINARY_PATH",
      "args": ["-db", "$DB_PATH"]
    }
  }
}
EOF
    echo -e "${GREEN}✓${NC} Configuration file created"
else
    echo -e "${YELLOW}Configuration file already exists${NC}"
    echo ""
    echo "Current configuration:"
    cat "$CONFIG_FILE"
    echo ""
    echo -e "${YELLOW}Please manually add the following to your mcpServers object:${NC}"
    echo ""
    echo "  \"moneywiz\": {"
    echo "    \"command\": \"$BINARY_PATH\","
    echo "    \"args\": [\"-db\", \"$DB_PATH\"]"
    echo "  }"
    echo ""
    echo "Edit the file at: $CONFIG_FILE"
fi

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. If Claude Desktop is running, quit it completely (⌘Q)"
echo "2. Reopen Claude Desktop"
echo "3. The MoneyWiz MCP server should connect automatically"
echo ""
echo "Test it by asking Claude:"
echo "  'Can you list my MoneyWiz accounts?'"
