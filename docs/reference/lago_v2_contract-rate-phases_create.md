## lago v2 contract-rate-phases create

Create rate phase

### Synopsis

Inserts a phase and renumbers later positions. The last phase must be indefinite and no earlier phase may be indefinite. Parent locking rules apply.
Requires the product catalog to be enabled and contract_rate_card write permission on the API key.

```
lago v2 contract-rate-phases create <contract_external_id> <applied_rate_card_code> [flags]
```

### Examples

```
  lago v2 contract-rate-phases create <contract_external_id> <applied_rate_card_code> --input @payload.json
  lago v2 contract-rate-phases create <contract_external_id> <applied_rate_card_code> --input @payload.json --output json  # full resource
```

### Options

```
      --billing-interval-cycle-count string   Number of billing cycles. Only the final phase is indefinite, represented by null.
      --code string                           Code unique within the applied rate card.
  -h, --help                                  help for create
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

* [lago v2 contract-rate-phases](lago_v2_contract-rate-phases)	 - Manage Lago contract rate phases
