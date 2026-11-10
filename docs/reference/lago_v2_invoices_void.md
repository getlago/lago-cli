## lago v2 invoices void

Void an invoice

### Synopsis

Voids an eligible invoice. Optional credit note creation requires a premium license.
Requires the product catalog to be enabled and invoice write permission on the API key.

```
lago v2 invoices void <lago_id> [flags]
```

### Examples

```
  lago v2 invoices void <lago_id> --input @payload.json
  lago v2 invoices void <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
      --credit-amount string          Requested credit amount, in cents.
      --generate-credit-note string   Whether to generate credit notes. Requires a premium license.
  -h, --help                          help for void
      --input string                  Complete JSON request body or @file.json
      --refund-amount string          Requested refund amount, in cents.
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

* [lago v2 invoices](lago_v2_invoices)	 - Manage Lago invoices
