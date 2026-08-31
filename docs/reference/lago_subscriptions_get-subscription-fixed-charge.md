## lago subscriptions get-subscription-fixed-charge

Retrieve a fixed charge for a subscription

### Synopsis

This endpoint retrieves a specific effective fixed charge for a subscription.
The `units` reflect any per-subscription unit override applied to the subscription; when no override exists, the plan-level units are returned.

```
lago subscriptions get-subscription-fixed-charge <external_id> <fixed_charge_code> [flags]
```

### Examples

```
  lago subscriptions get-subscription-fixed-charge <external_id> <fixed_charge_code>
```

### Options

```
  -h, --help                         help for get-subscription-fixed-charge
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
