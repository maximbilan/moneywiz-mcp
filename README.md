# MoneyWiz MCP Server

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

1. Clone this repository:
```bash
git clone <repository-url>
cd moneywiz-mcp
```

2. Build the server:
```bash
go build -o moneywiz-mcp ./cmd/main.go
```

3. Export your MoneyWiz database:
   - Open the MoneyWiz app
   - Go to **Settings** → **Database & Export** → **Export database file**
   - This will create a folder (e.g., `iMoneyWiz-Data-Backup-2025_12_21-17_23`) containing the SQLite database file (`ipadMoneyWiz.sqlite`)

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

**Note**: You need to export the database from MoneyWiz first (Settings → Database & Export → Export database file) to get the SQLite file.

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

```json
{
  "mcpServers": {
    "moneywiz": {
      "command": "moneywiz-mcp/moneywiz-mcp",
      "args": ["-db", "moneywiz-mcp/iMoneyWiz-Data-Backup-2025_12_21-17_23"]
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

### Dependencies

- `github.com/mark3labs/mcp-go` - MCP server library
- `github.com/mattn/go-sqlite3` - SQLite driver

## License

See LICENSE file for details.