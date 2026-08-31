## lago payment-methods list-customers

List all customer payment methods

### Synopsis

This endpoint retrieves all payment methods of a Customer.

```
lago payment-methods list-customers <external_customer_id> [flags]
```

### Examples

```
  lago payment-methods list-customers <external_customer_id>
```

### Options

```
      --all                       Fetch every page
  -h, --help                      help for list-customers
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

* [lago payment-methods](lago_payment-methods)	 - Manage Lago payment methods
