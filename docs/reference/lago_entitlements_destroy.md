## lago entitlements destroy

Delete an entitlement

### Synopsis

This endpoint deletes an existing entitlement by removing the feature from the plan.

```
lago entitlements destroy <code> <feature_code> [flags]
```

### Examples

```
  lago entitlements destroy <code> <feature_code>
```

### Options

```
  -h, --help   help for destroy
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
