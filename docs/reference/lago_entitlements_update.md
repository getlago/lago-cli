## lago entitlements update

Partial update of an entitlement

### Synopsis

This accepts a list of entitlements to update. If the feature isn't part of the plan yet, it's added with all the privileges from the payload. If the feature is already part of the plan, the privilege and values are updated or added. All privileges must be valid for the feature. All features  and privileges not part of the payload are left untouched. To remove privileges or features, use the DELETE endpoints.

```
lago entitlements update <code> [flags]
```

### Examples

```
  lago entitlements update <code> --input @payload.json
  lago entitlements update <code> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help                     help for update
      --idempotency-key string   Idempotency key for safe mutation retries
      --input string             Complete JSON request body or @file.json
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
