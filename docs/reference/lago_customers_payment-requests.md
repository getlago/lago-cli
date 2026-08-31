## lago customers payment-requests

List all customer's payment requests

### Synopsis

This endpoint is used to list all existing payment requests of a customer.

```
lago customers payment-requests <external_customer_id> [flags]
```

### Examples

```
  lago customers payment-requests <external_customer_id>
```

### Options

```
      --all                           Fetch every page
      --billing-entity-codes string   Filter payment requests by billing entity codes.
  -h, --help                          help for payment-requests
      --limit string                  Maximum number of results
      --page string                   Page number.
      --payment-status string         Filter by payment status. Possible values are 'pending', 'failed' or 'succeeded'.; one of: pending, failed, succeeded
      --per-page string               Number of records per page.
      --watch                         Poll and re-render when the response changes
      --watch-interval duration       Polling interval used with --watch (default 2s)
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

* [lago customers](lago_customers)	 - Manage Lago customers
