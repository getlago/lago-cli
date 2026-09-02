## lago customers applied-coupons

List all customer's applied coupons

### Synopsis

This endpoint is used to list all applied coupons for a customer.

```
lago customers applied-coupons <external_customer_id> [flags]
```

### Examples

```
  lago customers applied-coupons <external_customer_id>
```

### Options

```
      --all                       Fetch every page
      --coupon-code string        The code of the coupon applied to the customer. Use it to filter applied coupons by their code.
  -h, --help                      help for applied-coupons
      --limit string              Results per page (1-1000)
      --page string               Page number.
      --per-page string           Number of records per page.
      --status string             The status of the coupon. Can be either 'active' or 'terminated'.; one of: active, terminated
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

* [lago customers](lago_customers)	 - Manage Lago customers
