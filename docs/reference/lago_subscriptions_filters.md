## lago subscriptions filters

List all filters for a charge

### Synopsis

This endpoint retrieves all filters for a specific charge on a subscription.

```
lago subscriptions filters <external_id> <charge_code> [flags]
```

### Examples

```
  lago subscriptions filters <external_id> <charge_code>
```

### Options

```
      --all                          Fetch every page
  -h, --help                         help for filters
      --limit string                 Results per page (1-1000)
      --page string                  Page number.
      --per-page string              Number of records per page.
      --subscription-status string   Filter by subscription status. When provided, the subscription is looked up with this status instead of the default 'active' status. Possible values are 'pending', 'active', 'terminated', or 'canceled'.; one of: pending, active, terminated, canceled
      --watch                        Poll and re-render when the response changes
      --watch-interval duration      Polling interval used with --watch (default 2s)
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

* [lago subscriptions](lago_subscriptions)	 - Manage Lago subscriptions
