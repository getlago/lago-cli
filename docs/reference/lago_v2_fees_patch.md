## lago v2 fees patch

Update fee

### Synopsis

PUT and PATCH both accept partial updates. Omitted fields are unchanged.
Updates only payment_status. An invoiced fee cannot be updated and returns 422 with invoiced_fee.
Requires the product catalog to be enabled and fee write permission on the API key.

```
lago v2 fees patch <lago_id> [flags]
```

### Examples

```
  lago v2 fees patch <lago_id> --input @payload.json
  lago v2 fees patch <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help                    help for patch
      --input string            Complete JSON request body or @file.json
      --payment-status string   Payment status of the fee.; one of: pending, succeeded, failed, refunded
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

* [lago v2 fees](lago_v2_fees)	 - Manage Lago fees
