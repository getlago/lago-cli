## lago v2 plan-rate-phases destroy

Delete rate phase

### Synopsis

Removes the phase and renumbers later positions. The final indefinite phase and the only remaining phase cannot be deleted.
Requires the product catalog to be enabled and plan_rate_card write permission on the API key.

```
lago v2 plan-rate-phases destroy <plan_code> <applied_rate_card_code> <code> [flags]
```

### Examples

```
  lago v2 plan-rate-phases destroy <plan_code> <applied_rate_card_code> <code>
  lago v2 plan-rate-phases destroy <plan_code> <applied_rate_card_code> <code> --output json  # full resource
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

* [lago v2 plan-rate-phases](lago_v2_plan-rate-phases)	 - Manage Lago plan rate phases
