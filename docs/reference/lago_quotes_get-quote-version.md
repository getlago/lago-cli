## lago quotes get-quote-version

Retrieve a quote version

### Synopsis

This endpoint retrieves a specific quote version, along with its document `content` and its `billing_items`.
This is a premium feature.

```
lago quotes get-quote-version <lago_id> [flags]
```

### Examples

```
  lago quotes get-quote-version <lago_id>
```

### Options

```
  -h, --help                      help for get-quote-version
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

* [lago quotes](lago_quotes)	 - Manage Lago quotes
