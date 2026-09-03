## lago plans destroy-plan-charge-filter

Delete a charge filter

### Synopsis

This endpoint deletes a specific filter from a charge.

```
lago plans destroy-plan-charge-filter <code> <charge_code> <filter_id> [flags]
```

### Examples

```
  lago plans destroy-plan-charge-filter <code> <charge_code> <filter_id> --input @payload.json
  lago plans destroy-plan-charge-filter <code> <charge_code> <filter_id> --input @payload.json --output json  # full resource
```

### Options

```
      --cascade-updates string   When set to 'true', the deletion will be cascaded to the children plans.
  -h, --help                     help for destroy-plan-charge-filter
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
