## lago payments create

Create a payment

### Synopsis

This endpoint is used to create a manual payment

```
lago payments create [flags]
```

### Examples

```
  lago payments create --input @payload.json
  lago payments create --input @payload.json --output json  # full resource
```

### Options

```
      --amount-cents string      The payment amount in cents
  -h, --help                     help for create
      --idempotency-key string   Idempotency key for safe mutation retries
      --input string             Complete JSON request body or @file.json
      --invoice-id string        Unique identifier assigned to the invoice
      --paid-at string           The date the payment was made
      --reference string         Reference for the payment
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

* [lago payments](lago_payments)	 - Manage Lago payments
