## lago wallets get-customer-wallet-alerts

List wallet alerts

### Synopsis

This endpoint enables the retrieval of all alerts for a wallet.

```
lago wallets get-customer-wallet-alerts <external_customer_id> <wallet_code> [flags]
```

### Examples

```
  lago wallets get-customer-wallet-alerts <external_customer_id> <wallet_code>
```

### Options

```
  -h, --help                      help for get-customer-wallet-alerts
      --watch                     Poll and re-render when the response changes
      --watch-interval duration   Polling interval used with --watch (default 2s)
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
