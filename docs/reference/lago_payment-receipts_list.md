## lago payment-receipts list

List all payment receipts

### Synopsis

This endpoint is used to list all existing payment receipts.

```
lago payment-receipts list [flags]
```

### Examples

```
  lago payment-receipts list
```

### Options

```
      --all                       Fetch every page
  -h, --help                      help for list
      --invoice-id string         Filter payment receipts by invoice id.
      --limit string              Maximum number of results
      --page string               Page number.
      --per-page string           Number of records per page.
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

* [lago payment-receipts](lago_payment-receipts)	 - Manage Lago payment receipts
