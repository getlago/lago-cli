## lago v2 plan-applied-rate-cards create

Create applied rate card

### Synopsis

Attaches a rate card in the same currency. Only one rate card may price each product/filter combination. Plans attached to contracts are locked.
Requires the product catalog to be enabled and plan_rate_card write permission on the API key.

```
lago v2 plan-applied-rate-cards create <plan_code> [flags]
```

### Examples

```
  lago v2 plan-applied-rate-cards create <plan_code> --input @payload.json
  lago v2 plan-applied-rate-cards create <plan_code> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help                    help for create
      --input string            Complete JSON request body or @file.json
      --rate-card-code string   Code of the rate card to attach.
      --rate-phases string      Complete phase sequence with positions exactly 1 through the number of phases. Omitted or null creates one default indefinite phase. An explicit empty list is rejected.
      --units string            Quantity for fixed pricing. Null uses the default quantity. Sent as a decimal string.
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
