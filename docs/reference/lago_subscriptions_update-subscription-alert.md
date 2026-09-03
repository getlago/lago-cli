## lago subscriptions update-subscription-alert

Update a subscription alert

### Synopsis

This endpoint allows you to update an existing alert for a subscription.

```
lago subscriptions update-subscription-alert <external_id> <code> [flags]
```

### Examples

```
  lago subscriptions update-subscription-alert <external_id> <code> --input @payload.json
  lago subscriptions update-subscription-alert <external_id> <code> --input @payload.json --output json  # full resource
```

### Options

```
      --billable-metric-code string   The code of the billable metric associated with the alert. Only for alerts based on a billable metric.
      --code string                   Unique code used to identify the alert.
  -h, --help                          help for update-subscription-alert
      --input string                  Complete JSON request body or @file.json
      --name string                   The name of the alert.
      --subscription-status string    Filter by subscription status. When provided, the subscription is looked up with this status instead of the default 'active' status. Possible values are 'pending', 'active', 'terminated', or 'canceled'.; one of: pending, active, terminated, canceled
      --thresholds string             Array of thresholds associated with the alert.
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
