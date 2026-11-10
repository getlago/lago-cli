## lago v2 plan-applied-rate-cards get

Retrieve applied rate card

### Synopsis

Returns the record without nested relationships unless a supported expansion is requested.
Requires the product catalog to be enabled and plan_rate_card read permission on the API key.

```
lago v2 plan-applied-rate-cards get <plan_code> <code> [flags]
```

### Examples

```
  lago v2 plan-applied-rate-cards get <plan_code> <code>
```

### Options

```
  -h, --help                      help for get
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

* [lago v2 plan-applied-rate-cards](lago_v2_plan-applied-rate-cards)	 - Manage Lago plan applied rate cards
