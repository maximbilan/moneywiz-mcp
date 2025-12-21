# Setting Up MoneyWiz MCP Server with ChatGPT/Claude Desktop

This guide will help you configure the MoneyWiz MCP server to work with your ChatGPT or Claude Desktop app.

## Prerequisites

1. ✅ Built the server binary (`moneywiz-mcp`)
2. ✅ Have the MoneyWiz database folder (`iMoneyWiz-Data-Backup-2025_12_21-17_23`)

## Step 1: Get Absolute Paths

First, you need to find the absolute paths to:
- The `moneywiz-mcp` binary
- The MoneyWiz database folder

Run these commands in your terminal:

```bash
# Get the absolute path to the binary
cd /Users/max/Developer/moneywiz-mcp
realpath moneywiz-mcp || echo "$(pwd)/moneywiz-mcp"

# Get the absolute path to the database folder
realpath iMoneyWiz-Data-Backup-2025_12_21-17_23 || echo "$(pwd)/iMoneyWiz-Data-Backup-2025_12_21-17_23"
```

**Note these paths** - you'll need them in the next step.

## Step 2: Configure Claude Desktop (Recommended)

Claude Desktop has native MCP support. Here's how to set it up:

### 2.1. Locate the Configuration File

The configuration file is located at:
```
~/Library/Application Support/Claude/claude_desktop_config.json
```

### 2.2. Create or Edit the Configuration

**If the file doesn't exist**, create it:

```bash
mkdir -p ~/Library/Application\ Support/Claude
touch ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

**Edit the file** (use your preferred editor or `nano`):

```bash
nano ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

### 2.3. Add the MoneyWiz MCP Server Configuration

Add the following JSON configuration. **Replace the paths** with your actual absolute paths from Step 1:

```json
{
  "mcpServers": {
    "moneywiz": {
      "command": "/Users/max/Developer/moneywiz-mcp/moneywiz-mcp",
      "args": ["-db", "/Users/max/Developer/moneywiz-mcp/iMoneyWiz-Data-Backup-2025_12_21-17_23"]
    }
  }
}
```

**Important**: 
- Use absolute paths (starting with `/`)
- Make sure the binary is executable: `chmod +x /Users/max/Developer/moneywiz-mcp/moneywiz-mcp`
- The `args` array should point to the **folder** containing `ipadMoneyWiz.sqlite`, not the file itself

### 2.4. If You Already Have Other MCP Servers

If you already have other MCP servers configured, add `moneywiz` to the existing `mcpServers` object:

```json
{
  "mcpServers": {
    "existing-server": {
      "command": "...",
      "args": [...]
    },
    "moneywiz": {
      "command": "/Users/max/Developer/moneywiz-mcp/moneywiz-mcp",
      "args": ["-db", "/Users/max/Developer/moneywiz-mcp/iMoneyWiz-Data-Backup-2025_12_21-17_23"]
    }
  }
}
```

### 2.5. Restart Claude Desktop

1. Quit Claude Desktop completely (⌘Q)
2. Reopen Claude Desktop
3. The MCP server should automatically connect

### 2.6. Verify It's Working

In Claude Desktop, you should see:
- A connection indicator showing the MCP server is connected
- The MoneyWiz tools available when you ask Claude to help with your finances

Try asking Claude:
> "Can you list my MoneyWiz accounts?"

## Step 3: Configure ChatGPT (If Applicable)

**Note**: As of now, ChatGPT's native app may not have MCP support. However, if you're using:

### Option A: ChatGPT with MCP via API/Plugin
Some integrations may support MCP. Check your specific ChatGPT integration documentation.

### Option B: Use Claude Desktop Instead
Claude Desktop has excellent MCP support and works similarly to ChatGPT. We recommend using Claude Desktop for testing.

## Troubleshooting

### Server Not Connecting

1. **Check the binary path is correct**:
   ```bash
   ls -la /Users/max/Developer/moneywiz-mcp/moneywiz-mcp
   ```

2. **Make sure the binary is executable**:
   ```bash
   chmod +x /Users/max/Developer/moneywiz-mcp/moneywiz-mcp
   ```

3. **Test the server manually**:
   ```bash
   /Users/max/Developer/moneywiz-mcp/moneywiz-mcp -db /Users/max/Developer/moneywiz-mcp/iMoneyWiz-Data-Backup-2025_12_21-17_23
   ```
   (It should start and wait for input - press Ctrl+C to stop)

4. **Check the database path**:
   ```bash
   ls -la /Users/max/Developer/moneywiz-mcp/iMoneyWiz-Data-Backup-2025_12_21-17_23/ipadMoneyWiz.sqlite
   ```

5. **Check Claude Desktop logs**:
   - Look for error messages in the Claude Desktop console
   - On macOS, you can check: `~/Library/Logs/Claude/`

### JSON Syntax Errors

Make sure your `claude_desktop_config.json` file is valid JSON:
- No trailing commas
- All strings in double quotes
- Properly closed braces

You can validate it with:
```bash
python3 -m json.tool ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

### Permission Issues

If you get permission errors:
```bash
chmod +x /Users/max/Developer/moneywiz-mcp/moneywiz-mcp
```

## Quick Test Script

Here's a quick script to set everything up:

```bash
#!/bin/bash

# Set your paths
BINARY_PATH="/Users/max/Developer/moneywiz-mcp/moneywiz-mcp"
DB_PATH="/Users/max/Developer/moneywiz-mcp/iMoneyWiz-Data-Backup-2025_12_21-17_23"

# Make binary executable
chmod +x "$BINARY_PATH"

# Create config directory if it doesn't exist
mkdir -p ~/Library/Application\ Support/Claude

# Create or update config file
CONFIG_FILE=~/Library/Application\ Support/Claude/claude_desktop_config.json

if [ ! -f "$CONFIG_FILE" ]; then
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
else
    echo "Config file exists. Please manually add the moneywiz server configuration."
    echo "Edit: $CONFIG_FILE"
fi

echo "Configuration complete! Restart Claude Desktop to apply changes."
```

Save this as `setup.sh`, make it executable (`chmod +x setup.sh`), and run it.

## Next Steps

Once configured, you can:
1. Ask Claude to list your accounts
2. Query account balances
3. View recent transactions
4. Explore your financial data

Example prompts:
- "Show me all my MoneyWiz accounts"
- "What's the balance of account ID 249?"
- "List my last 10 transactions"
- "Show me all my expense categories"
