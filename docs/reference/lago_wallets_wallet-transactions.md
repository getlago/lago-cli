## lago wallets wallet-transactions

List all wallet transactions

### Synopsis

This endpoint is used to list all wallet transactions.

```
lago wallets wallet-transactions <lago_id> [flags]
```

### Examples

```
  lago wallets wallet-transactions <lago_id>
```

### Options

```
      --all                         Fetch every page
  -h, --help                        help for wallet-transactions
      --limit string                Results per page (1-1000)
      --page string                 Page number.
      --per-page string             Number of records per page.
      --status string               The status of the wallet transaction. Possible values are 'pending' or 'settled'.
      --transaction-status string   The transaction status of the wallet transaction. Possible values are 'purchased' (with pending or settled status), 'granted' (without invoice_id), 'voided' or 'invoiced'.
      --transaction-type string     The transaction type of the wallet transaction. Possible values are 'inbound' (increasing the wallet balance) or 'outbound' (decreasing the wallet balance).
      --watch                       Poll and re-render when the response changes
      --watch-interval duration     Polling interval used with --watch (default 2s)
```

### Options inherited from parent commands

```
      --api-key string     Override the Lago API key
      --api-url string     Override the Lago API URL
      --confirm string     Confirm a dangerous operation with its resource identifier
      --dry-run            Print mutating requests without sending them
      --insecure           Allow insecure HTTP or TLS for self-hosted Lago
      --mode string        Environment mode: live or test
      --no-retry           Disable automatic retries
  -o, --output string      Output format: table, json, or yaml (default "table")
      --profile string     Named profile to use
      --query string       JMESPath expression applied to the response
      --timeout duration   Total request timeout (default 30s)
      --timing             Print request latency breakdown
      --verbose            Print redacted request and response details
```

### SEE ALSO

* [lago wallets](lago_wallets)	 - Manage Lago wallets
