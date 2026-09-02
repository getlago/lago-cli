## lago coupons list-applied

List all applied coupons

### Synopsis

This endpoint is used to list all applied coupons. You can filter by coupon status and by customer.

```
lago coupons list-applied [flags]
```

### Examples

```
  lago coupons list-applied
```

### Options

```
      --all                           Fetch every page
      --coupon-code string            The code of the coupon applied to the customer. Use it to filter applied coupons by their code.
      --external-customer-id string   The customer external unique identifier (provided by your own application)
  -h, --help                          help for list-applied
      --limit string                  Results per page (1-1000)
      --page string                   Page number.
      --per-page string               Number of records per page.
      --status string                 The status of the coupon. Can be either 'active' or 'terminated'.; one of: active, terminated
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

* [lago coupons](lago_coupons)	 - Manage Lago coupons
