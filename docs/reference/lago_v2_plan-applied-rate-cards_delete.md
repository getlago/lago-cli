## lago v2 plan-applied-rate-cards delete

Delete applied rate card

### Synopsis

Removes the rate card from an editable plan.
Requires the product catalog to be enabled and plan_rate_card write permission on the API key.

```
lago v2 plan-applied-rate-cards delete <plan_code> <code> [flags]
```

### Examples

```
  lago v2 plan-applied-rate-cards delete <plan_code> <code>
  lago v2 plan-applied-rate-cards delete <plan_code> <code> --output json  # full resource
```

### Options

```
  -h, --help   help for delete
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

* [lago v2 plan-applied-rate-cards](lago_v2_plan-applied-rate-cards)	 - Manage Lago plan applied rate cards
