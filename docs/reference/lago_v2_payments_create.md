## lago v2 payments create

Create payment

### Synopsis

Records a manual payment against an existing invoice. A reference and positive integer amount are required. Premium access and invoice eligibility are checked, except for supported advance-fee payments.
Requires the product catalog to be enabled and payment write permission on the API key.

```
lago v2 payments create [flags]
```

### Examples

```
  lago v2 payments create --input @payload.json
  lago v2 payments create --input @payload.json --output json  # full resource
```

### Options

```
      --amount-cents string   Amount paid, in cents.
  -h, --help                  help for create
      --input string          Complete JSON request body or @file.json
      --invoice-id string     Identifier of the invoice being paid.
      --paid-at string        Payment timestamp. Defaults to now.
      --reference string      Manual payment reference.
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

* [lago v2 payments](lago_v2_payments)	 - Manage Lago payments
