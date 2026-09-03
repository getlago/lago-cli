## lago entitlements destroy-subscription-entitlement-privilege

Remove a privilege from a subscription entitlement override

### Synopsis

This endpoint removes a specific privilege from a subscription entitlement. The privilege entitlement remains available from the plan.

```
lago entitlements destroy-subscription-entitlement-privilege <external_id> <feature_code> <privilege_code> [flags]
```

### Examples

```
  lago entitlements destroy-subscription-entitlement-privilege <external_id> <feature_code> <privilege_code>
```

### Options

```
  -h, --help                         help for destroy-subscription-entitlement-privilege
      --subscription-status string   Filter by subscription status. When provided, the subscription is looked up with this status instead of the default 'active' status. Possible values are 'pending', 'active', 'terminated', or 'canceled'.; one of: pending, active, terminated, canceled
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

* [lago entitlements](lago_entitlements)	 - Manage Lago entitlements
