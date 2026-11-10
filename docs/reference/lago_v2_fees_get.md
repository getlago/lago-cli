## lago v2 fees get

Retrieve fee

### Synopsis

Legacy fees return 404. Related products, filters and rates can identify deleted catalog records for historical fees.
Requires the product catalog to be enabled and fee read permission on the API key.

```
lago v2 fees get <lago_id> [flags]
```

### Examples

```
  lago v2 fees get <lago_id>
```

### Options

```
      --expands string            Relationships to include in this response. Only one expansion level is supported.
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

* [lago v2 fees](lago_v2_fees)	 - Manage Lago fees
