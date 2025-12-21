# MoneyWiz MCP Server

An MCP (Model Context Protocol) server for accessing MoneyWiz database data. This server allows ChatGPT, Claude, and other MCP-compatible clients to query your MoneyWiz financial data.

## Features

- **List Accounts**: Get all accounts with balances and currencies
- **Get Account Balance**: Retrieve balance for a specific account
- **List Transactions**: View recent transactions, optionally filtered by account
- **List Categories**: Get all expense/income categories
- **List Budgets**: Get all budgets (if available)

## Installation

1. Clone this repository:
```bash
git clone <repository-url>
cd moneywiz-mcp
```

2. Build the server:
```bash
go build -o moneywiz-mcp ./cmd/main.go
```

## Usage

### Running the Server

The server can be run with the database path specified via command-line argument:

```bash
./moneywiz-mcp -db /path/to/iMoneyWiz-Data-Backup-2025_12_21-17_23
```

Or if the database folder is in the current directory:

```bash
./moneywiz-mcp -db ./iMoneyWiz-Data-Backup-2025_12_21-17_23
```

The server will automatically look for `ipadMoneyWiz.sqlite` in the specified folder.

### MCP Client Configuration

To use this server with an MCP client (like Claude Desktop), add it to your MCP configuration file.

#### Quick Setup (Recommended)

Run the setup script:

```bash
./setup.sh
```

This will automatically configure Claude Desktop with the correct paths.

#### Manual Configuration

**Claude Desktop Configuration**

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

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

**Important**: Replace the paths with your actual absolute paths.

After updating the configuration:
1. Quit Claude Desktop completely (⌘Q)
2. Reopen Claude Desktop
3. The MCP server should connect automatically

For detailed setup instructions, see [SETUP.md](SETUP.md).

## Available Tools

### `list_accounts`

List all accounts in MoneyWiz with their balances and currencies.

**Parameters**: None

**Example**:
```json
{
  "name": "list_accounts",
  "arguments": {}
}
```

### `get_account_balance`

Get the balance for a specific account by ID.

**Parameters**:
- `account_id` (integer, required): The ID of the account

**Example**:
```json
{
  "name": "get_account_balance",
  "arguments": {
    "account_id": 249
  }
}
```

### `list_transactions`

List recent transactions, optionally filtered by account ID.

**Parameters**:
- `account_id` (integer, optional): Account ID to filter transactions. If not provided, returns all transactions
- `limit` (integer, optional): Maximum number of transactions to return (default: 50)

**Example**:
```json
{
  "name": "list_transactions",
  "arguments": {
    "account_id": 249,
    "limit": 20
  }
}
```

### `list_categories`

List all categories in MoneyWiz.

**Parameters**: None

**Example**:
```json
{
  "name": "list_categories",
  "arguments": {}
}
```

### `list_budgets`

List all budgets in MoneyWiz.

**Parameters**: None

**Example**:
```json
{
  "name": "list_budgets",
  "arguments": {}
}
```

## Database Structure

This server accesses the MoneyWiz SQLite database (`ipadMoneyWiz.sqlite`). The database uses Core Data's entity-attribute-value model, where most objects are stored in the `ZSYNCOBJECT` table with different entity types (`Z_ENT`):
- Entity 16: Accounts
- Entity 34: Transactions
- Entity 19: Categories
- Entity 18: Budgets

## Development

### Project Structure

```
moneywiz-mcp/
├── cmd/
│   └── main.go          # Main entry point
├── internal/
│   ├── database/        # Database access layer
│   │   └── database.go
│   └── server/          # MCP server implementation
│       └── server.go
├── go.mod
└── README.md
```

### Building

```bash
go build -o moneywiz-mcp ./cmd/main.go
```

### Dependencies

- `github.com/mark3labs/mcp-go` - MCP server library
- `github.com/mattn/go-sqlite3` - SQLite driver

## License

See LICENSE file for details.

## Notes

- This is a basic implementation focusing on read-only operations
- The database structure is based on Core Data, which can be complex
- Some features may not work if your MoneyWiz database structure differs
- Always backup your MoneyWiz database before using this tool