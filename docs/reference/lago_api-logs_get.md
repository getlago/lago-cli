## lago api-logs get

Retrieve an api log

### Synopsis

This endpoint retrieves an existing api log that represents a request made to the API. The api log is identified by its unique request_id.

```
lago api-logs get <request_id> [flags]
```

### Examples

```
  lago api-logs get <request_id>
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

* [lago api-logs](lago_api-logs)	 - Manage Lago api logs
