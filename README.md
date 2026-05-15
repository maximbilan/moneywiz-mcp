# MoneyWiz MCP Server

[![Build and Test](https://github.com/maximbilan/moneywiz-mcp/actions/workflows/build-test.yml/badge.svg)](https://github.com/maximbilan/moneywiz-mcp/actions/workflows/build-test.yml)
[![Release](https://img.shields.io/github/v/release/maximbilan/moneywiz-mcp)](https://github.com/maximbilan/moneywiz-mcp/releases)

An MCP (Model Context Protocol) server for accessing MoneyWiz database data. This server allows ChatGPT, Claude, and other MCP-compatible clients to query your MoneyWiz financial data, analyze spending and income trends, get savings recommendations, calculate net worth, and view comprehensive financial statistics.

## Features

- **List Accounts**: Get all accounts with balances and currencies
- **Get Account Balance**: Retrieve balance for a specific account
- **List Transactions**: View recent transactions, optionally filtered by account
- **List Categories**: Get all expense/income categories
- **Analyze Spending Trends**: Analyze spending trends by category and time period (month/year)
- **Analyze Income Trends**: Analyze income trends by category and time period (month/year)
- **Get Savings Recommendations**: Get personalized savings recommendations based on income vs spending
- **Calculate Net Worth**: Calculate total net worth from all accounts (assets minus liabilities)
- **Get Financial Stats**: Get comprehensive financial statistics from all historical data

## Installation

### Prerequisites

- macOS (for MoneyWiz export + Claude Desktop path conventions)
- Go 1.22+ (required when installing from source)
- MoneyWiz exported database folder containing `ipadMoneyWiz.sqlite`

### 1. Clone

```bash
git clone <repository-url>
cd moneywiz-mcp
```

### 2. One-click install (recommended)

This builds the binary, installs it to `~/.local/bin/moneywiz-mcp`, and registers it in Claude clients.
If you omit `--db`, it reuses `~/.moneywiz-mcp/ipadMoneyWiz.sqlite` or imports the newest local `iMoneyWiz-Data-Backup-*` export automatically.

```bash
./scripts/install.sh
```

With explicit database path (folder/sqlite):

```bash
./scripts/install.sh --db "/path/to/iMoneyWiz-Data-Backup-2025_12_21-17_23"
```

```bash
./scripts/install.sh --db latest
```

### Release downloads

Tagged releases publish:
- `moneywiz-mcp-darwin-arm64`
- `moneywiz-mcp-darwin-amd64`
- `moneywiz-mcp-darwin-universal`
- `SHA256SUMS`

The release workflow does not publish a standalone installer yet. `scripts/install.sh` is source-tree based and is meant to be run from this repository.

### 3. Import DB to a stable path (recommended)

Avoid dealing with changing export folder names by importing to:
`~/.moneywiz-mcp/ipadMoneyWiz.sqlite`

```bash
./scripts/import_db.sh "/path/to/iMoneyWiz-Data-Backup-2025_12_21-17_23"
```

Then install using the stable file:

```bash
./scripts/install.sh --db "$HOME/.moneywiz-mcp/ipadMoneyWiz.sqlite"
```

Or let the installer do the import automatically:

```bash
./scripts/install.sh
```

### 4. One-click rebuild + reinstall (debug/dev loop)

Use this after local code changes. It rebuilds and re-registers in one command.

```bash
./scripts/rebuild_reinstall.sh --db "/path/to/iMoneyWiz-Data-Backup-2025_12_21-17_23"
```

### 5. Manual build (optional)

```bash
go build -o moneywiz-mcp ./cmd/main.go
```

### 6. Uninstall

Remove MCP registration from Claude clients and delete installed binary:

```bash
./scripts/uninstall.sh
```

Non-interactive:

```bash
./scripts/uninstall.sh --yes
```

### Export MoneyWiz Database

1. Open MoneyWiz
2. Go to **Settings** → **Database & Export** → **Export database file**
3. Use the exported folder that contains `ipadMoneyWiz.sqlite`

## Usage

### Running the Server

Run directly with a database folder or sqlite file:

```bash
./moneywiz-mcp -db /path/to/iMoneyWiz-Data-Backup-2025_12_21-17_23
```

The binary accepts either:
- Export folder path (it appends `ipadMoneyWiz.sqlite`)
- Direct sqlite file path
- `latest` (pick newest `iMoneyWiz-Data-Backup-*` found)
- No `-db` (priority order below)

Path resolution priority:
1. `-db` argument
2. `MONEYWIZ_DB_PATH` env var
3. `~/.moneywiz-mcp/ipadMoneyWiz.sqlite` if present
4. Auto-detect newest export folder in common locations

### MCP Client Configuration

`./scripts/install.sh` already handles configuration for:
- Claude Desktop (`claude_desktop_config.json`)
- Claude Code (`claude mcp add`)

#### Manual Configuration

**Claude Desktop Configuration**

```json
{
  "mcpServers": {
    "moneywiz": {
      "command": "/Users/<you>/.local/bin/moneywiz-mcp",
      "args": ["-db", "/absolute/path/to/ipadMoneyWiz.sqlite"]
    }
  }
}
```

**Claude Code (manual)**

```bash
claude mcp add --scope user moneywiz -- /Users/<you>/.local/bin/moneywiz-mcp -db /absolute/path/to/ipadMoneyWiz.sqlite
```

**Important**: Use absolute paths.

After updating the configuration:
1. Quit Claude Desktop completely (⌘Q)
2. Reopen Claude Desktop
3. The MCP server should connect automatically

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

### `analyze_spending_trends`

Analyze spending trends by category and time period. Groups spending by month or year and provides category breakdowns.

**Parameters**:
- `group_by` (string, optional): Group by `"month"` or `"year"` (default: `"month"`)
- `months` (integer, optional): Number of months to analyze (default: 6)

**Example**:
```json
{
  "name": "analyze_spending_trends",
  "arguments": {
    "group_by": "month",
    "months": 6
  }
}
```

**Returns**: Array of spending trends with:
- `period`: Time period (YYYY-MM or YYYY)
- `total_spending`: Total spending for the period
- `transaction_count`: Number of transactions
- `by_category`: Map of category names to spending amounts

### `analyze_income_trends`

Analyze income trends by category and time period. Groups income by month or year and provides category breakdowns.

**Parameters**:
- `group_by` (string, optional): Group by `"month"` or `"year"` (default: `"month"`)
- `months` (integer, optional): Number of months to analyze (default: 6)

**Example**:
```json
{
  "name": "analyze_income_trends",
  "arguments": {
    "group_by": "year",
    "months": 12
  }
}
```

**Returns**: Array of income trends with:
- `period`: Time period (YYYY-MM or YYYY)
- `total_income`: Total income for the period
- `transaction_count`: Number of transactions
- `by_category`: Map of category names to income amounts

### `get_savings_recommendations`

Analyze income vs spending and get personalized savings recommendations. Provides actionable advice based on your financial patterns.

**Parameters**:
- `months` (integer, optional): Number of months to analyze (default: 6)

**Example**:
```json
{
  "name": "get_savings_recommendations",
  "arguments": {
    "months": 6
  }
}
```

**Returns**: Savings analysis with:
- `period`: Analysis period description
- `total_income`: Total income for the period
- `total_spending`: Total spending for the period
- `net_savings`: Net savings (income - spending)
- `savings_rate`: Savings rate as percentage
- `average_monthly_income`: Average monthly income
- `average_monthly_spending`: Average monthly spending
- `top_spending_categories`: Top 5 spending categories with percentages
- `recommendations`: Array of recommendations with:
  - `type`: `"warning"`, `"suggestion"`, or `"positive"`
  - `title`: Recommendation title
  - `description`: Detailed recommendation
  - `priority`: `"high"`, `"medium"`, or `"low"`
  - `impact`: Potential savings amount

### `calculate_net_worth`

Calculate total net worth from all accounts. Sums all account balances (assets minus liabilities).

**Parameters**: None

**Example**:
```json
{
  "name": "calculate_net_worth",
  "arguments": {}
}
```

**Returns**: Net worth calculation with:
- `total_assets`: Sum of all positive account balances
- `total_liabilities`: Sum of all negative account balances (as positive values)
- `net_worth`: Total assets minus total liabilities
- `account_count`: Number of accounts included
- `by_currency`: Net worth broken down by currency
- `accounts`: Array of all accounts with balances

### `get_financial_stats`

Get comprehensive financial statistics from all historical data. Provides overview metrics and yearly breakdowns.

**Parameters**: None

**Example**:
```json
{
  "name": "get_financial_stats",
  "arguments": {}
}
```

**Returns**: Financial statistics with:
- `total_transactions`: Total number of transactions (all time)
- `income_transactions`: Number of income transactions
- `expense_transactions`: Number of expense transactions
- `total_income`: Total income (all time)
- `total_spending`: Total spending (all time)
- `net_savings`: Net savings (all time)
- `average_transaction`: Average transaction amount
- `largest_income`: Largest single income transaction
- `largest_expense`: Largest single expense transaction
- `account_count`: Total number of accounts
- `category_count`: Total number of categories
- `first_transaction_date`: Date of first transaction
- `last_transaction_date`: Date of last transaction
- `date_range`: Formatted date range string
- `by_year`: Map of yearly statistics with:
  - `year`: Year (YYYY)
  - `income`: Total income for the year
  - `spending`: Total spending for the year
  - `net_savings`: Net savings for the year
  - `transaction_count`: Number of transactions for the year

## Database Structure

This server accesses the MoneyWiz SQLite database (`ipadMoneyWiz.sqlite`). The database uses Core Data's entity-attribute-value model, where most objects are stored in the `ZSYNCOBJECT` table with different entity types (`Z_ENT`):
- Entity 10, 11, 12, 13, 15, 16: Accounts (various types)
- Entity 37, 45, 46, 47: Regular transactions
- Entity 43: Transfer transactions
- Entity 19: Categories

### Important Notes

- **Dates**: Transaction dates are stored as Core Data timestamps (seconds since 2001-01-01 UTC) and are automatically converted to ISO format
- **Balances**: Account balances are stored in `ZBALLANCE` (note the double L). If balance is 0 or NULL, it's calculated from opening balance + transactions
- **Transactions**: Income transactions have positive `ZAMOUNT1`, expense transactions have negative `ZAMOUNT1`
- **Categories**: Categories are linked to transactions via the `ZCATEGORYASSIGMENT` table

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

### Validation

```bash
gofmt -l .
go build ./...
go vet ./...
go test ./... -race
bash -n scripts/*.sh
```

### Release Validation

Before tagging a release:

```bash
gofmt -l .
go build ./...
go vet ./...
go test ./... -race
bash -n scripts/*.sh
```

### Dependencies

- `github.com/mark3labs/mcp-go` - MCP server library
- `github.com/mattn/go-sqlite3` - SQLite driver

## License

See [LICENSE](LICENSE) file for details.
