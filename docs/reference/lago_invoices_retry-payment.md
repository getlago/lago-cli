## lago invoices retry-payment

Retry an invoice payment

### Synopsis

This endpoint resends an invoice for collection and retry a payment.

```
lago invoices retry-payment <lago_id> [flags]
```

### Examples

```
  lago invoices retry-payment <lago_id> --input @payload.json
```

### Options

```
  -h, --help                         help for retry-payment
      --input string                 Complete JSON request body or @file.json
      --payment-method-id string     The unique identifier of the payment method (required when using a specific provider payment method).
      --payment-method-type string   The type of payment method to use.; one of: provider, manual
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

* [lago invoices](lago_invoices)	 - Manage Lago invoices
