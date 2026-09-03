## lago entitlements create

Create an entitlement

### Synopsis

This endpoint creates new entitlements by adding features to a plan. Note that all existing entitlements will be deleted and replaced by the ones provided. To add a new entitlement without removing the existing ones, use PATCH. The feature must exist and all privileges must be valid for the feature.

```
lago entitlements create <code> [flags]
```

### Examples

```
  lago entitlements create <code> --input @payload.json
  lago entitlements create <code> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help           help for create
      --input string   Complete JSON request body or @file.json
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
