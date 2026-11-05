## lago wallets update-customer-wallet-alert

Update a wallet alert

### Synopsis

This endpoint allows you to update an existing alert for a wallet. A non-empty `thresholds` array replaces the whole set of thresholds on the alert. When `thresholds` is left out, or sent as an empty array, the thresholds already on the alert are kept. A threshold sent with a `code` it already had, and without `notify_on`, keeps the `notify_on` it had.

```
lago wallets update-customer-wallet-alert <external_customer_id> <wallet_code> <code> [flags]
```

### Examples

```
  lago wallets update-customer-wallet-alert <external_customer_id> <wallet_code> <code> --input @payload.json
  lago wallets update-customer-wallet-alert <external_customer_id> <wallet_code> <code> --input @payload.json --output json  # full resource
```

### Options

```
      --code string         Unique code used to identify the alert.
  -h, --help                help for update-customer-wallet-alert
      --input string        Complete JSON request body or @file.json
      --name string         The name of the alert.
      --thresholds string   Array of thresholds associated with the alert. An alert accepts at most 20 thresholds, and a larger array returns a '422' error with 'too_many_thresholds' on 'thresholds'. Two thresholds cannot share the same value and the same 'recurring' flag. The value of a recurring threshold must be positive.
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
