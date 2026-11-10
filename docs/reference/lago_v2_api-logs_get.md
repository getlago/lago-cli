## lago v2 api-logs get

Retrieve api log

### Synopsis

Requires a premium license and available log storage. No expansions are supported.
Requires the product catalog to be enabled and api_log read permission on the API key.

```
lago v2 api-logs get <request_id> [flags]
```

### Examples

```
  lago v2 api-logs get <request_id>
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

* [lago v2 api-logs](lago_v2_api-logs)	 - Manage Lago api logs
