## lago plans destroy-plan-fixed-charge

Delete a fixed charge

### Synopsis

This endpoint deletes a specific fixed charge from a plan.

```
lago plans destroy-plan-fixed-charge <code> <fixed_charge_code> [flags]
```

### Examples

```
  lago plans destroy-plan-fixed-charge <code> <fixed_charge_code> --input @payload.json
```

### Options

```
      --cascade-updates string   When set to 'true', the deletion will be cascaded to the children plans.
  -h, --help                     help for destroy-plan-fixed-charge
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

* [lago plans](lago_plans)	 - Manage Lago plans
