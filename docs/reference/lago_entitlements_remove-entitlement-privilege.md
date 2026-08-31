## lago entitlements remove-entitlement-privilege

Remove a privilege from an entitlement

### Synopsis

This endpoint removes a specific privilege and its value from an entitlement. The privilege remains untouched on the original feature.

```
lago entitlements remove-entitlement-privilege <code> <feature_code> <privilege_code> [flags]
```

### Examples

```
  lago entitlements remove-entitlement-privilege <code> <feature_code> <privilege_code>
```

### Options

```
  -h, --help   help for remove-entitlement-privilege
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
