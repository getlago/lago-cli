## lago plans merge-plan-metadata

Merge plan metadata

### Synopsis

This endpoint merges the provided metadata with existing metadata on the plan.
Existing keys not in the request are preserved. New keys are added, existing keys are updated.

```
lago plans merge-plan-metadata <code> [flags]
```

### Examples

```
  lago plans merge-plan-metadata <code> --input @payload.json
```

### Options

```
  -h, --help           help for merge-plan-metadata
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

* [lago plans](lago_plans)	 - Manage Lago plans
