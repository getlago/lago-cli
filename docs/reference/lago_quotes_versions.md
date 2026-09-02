## lago quotes versions

List all versions of a quote

### Synopsis

This endpoint is used to list all the versions of a specific quote, from the most recent to the oldest.
The `content` and `billing_items` of each version are omitted; retrieve a version to get them.
This is a premium feature.

```
lago quotes versions <lago_id> [flags]
```

### Examples

```
  lago quotes versions <lago_id>
```

### Options

```
      --all                       Fetch every page
  -h, --help                      help for versions
      --limit string              Results per page (1-1000)
      --page string               Page number.
      --per-page string           Number of records per page.
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
