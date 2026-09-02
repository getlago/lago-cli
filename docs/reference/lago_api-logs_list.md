## lago api-logs list

List all api logs

### Synopsis

This endpoint retrieves all existing api logs that represent requests performed to Lago's API.

```
lago api-logs list [flags]
```

### Examples

```
  lago api-logs list
```

### Options

```
      --all                       Fetch every page
      --api-version string        Filter results by API version
      --from-date string          Filter api logs from a specific date.
  -h, --help                      help for list
      --http-methods string       Filter results by HTTP methods
      --http-statuses string      Filter results by HTTP status or by generic request status
      --limit string              Results per page (1-1000)
      --page string               Page number.
      --per-page string           Number of records per page.
      --request-paths string      Filter results by the path of the request
      --to-date string            Filter api logs up to a specific date.
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
