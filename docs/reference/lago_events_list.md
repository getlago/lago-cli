## lago events list

List all events

### Synopsis

This endpoint is used for retrieving all events.

```
lago events list [flags]
```

### Examples

```
  lago events list
```

### Options

```
      --all                                Fetch every page
      --code string                        Filter events by its code.
      --external-subscription-id string    External subscription ID
  -h, --help                               help for list
      --limit string                       Results per page (1-1000)
      --page string                        Page number.
      --per-page string                    Number of records per page.
      --timestamp-from string              Filter events by timestamp starting from a specific date.
      --timestamp-from-started-at string   Requires 'external_subscription_id' to be set. Filter events by timestamp after the subscription started at datetime.
      --timestamp-to string                Filter events by timestamp up to a specific date.
      --watch                              Poll and re-render when the response changes
      --watch-interval duration            Polling interval used with --watch (default 2s)
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

* [lago events](lago_events)	 - Manage Lago events
