## lago billable-metrics get

Retrieve a billable metric

### Synopsis

This endpoint retrieves an existing billable metric that represents a pricing component of your application. The billable metric is identified by its unique code.

```
lago billable-metrics get <code> [flags]
```

### Examples

```
  lago billable-metrics get <code>
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

* [lago billable-metrics](lago_billable-metrics)	 - Manage Lago billable metrics
