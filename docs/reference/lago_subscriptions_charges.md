## lago subscriptions charges

List all charges for a subscription

### Synopsis

This endpoint retrieves all effective charges for a specific subscription.
If the subscription has plan overrides with charge overrides, the overridden charges are returned.

```
lago subscriptions charges <external_id> [flags]
```

### Examples

```
  lago subscriptions charges <external_id>
```

### Options

```
      --all                          Fetch every page
  -h, --help                         help for charges
      --limit string                 Maximum number of results
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
