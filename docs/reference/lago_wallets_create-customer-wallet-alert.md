## lago wallets create-customer-wallet-alert

Create wallet alert(s)

### Synopsis

This endpoint allows you to create new alerts for a wallet. Send a single alert object wrapped in `alert` key to create one alert, or an array of alert objects wrapped in `alerts` key to create multiple alerts atomically.

```
lago wallets create-customer-wallet-alert <external_customer_id> <wallet_code> [flags]
```

### Examples

```
  lago wallets create-customer-wallet-alert <external_customer_id> <wallet_code> --input @payload.json
  lago wallets create-customer-wallet-alert <external_customer_id> <wallet_code> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help                     help for create-customer-wallet-alert
      --idempotency-key string   Idempotency key for safe mutation retries
      --input string             Complete JSON request body or @file.json
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
