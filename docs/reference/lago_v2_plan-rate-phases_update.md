## lago v2 plan-rate-phases update

Update rate phase

### Synopsis

PUT and PATCH both accept partial updates. Omitted fields are unchanged.
A null rate_override removes the override. Structural rate card fields are not overridable. Parent locking rules apply.
Requires the product catalog to be enabled and plan_rate_card write permission on the API key.

```
lago v2 plan-rate-phases update <plan_code> <applied_rate_card_code> <code> [flags]
```

### Examples

```
  lago v2 plan-rate-phases update <plan_code> <applied_rate_card_code> <code> --input @payload.json
  lago v2 plan-rate-phases update <plan_code> <applied_rate_card_code> <code> --input @payload.json --output json  # full resource
```

### Options

```
      --billing-interval-cycle-count string   Number of billing cycles. Only the final phase is indefinite, represented by null.
      --code string                           Code unique within the applied rate card.
  -h, --help                                  help for update
      --input string                          Complete JSON request body or @file.json
      --name string                           Name of the phase.
      --position string                       One-based sequence position.
      --rate-override string                  Pricing override. On update, null removes the existing override.
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

* [lago v2 plan-rate-phases](lago_v2_plan-rate-phases)	 - Manage Lago plan rate phases
