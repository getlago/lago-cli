## lago v2 rate-card-rates list

List rate card rates

### Synopsis

Returns the complete list without cursor metadata.
Rates are ordered by effective_from, newest first. Offset pagination parameters are rejected.
Requires the product catalog to be enabled and rate_card read permission on the API key.

```
lago v2 rate-card-rates list <rate_card_code> [flags]
```

### Examples

```
  lago v2 rate-card-rates list <rate_card_code>
```

### Options

```
  -h, --help                      help for list
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

* [lago v2 rate-card-rates](lago_v2_rate-card-rates)	 - Manage Lago rate card rates
