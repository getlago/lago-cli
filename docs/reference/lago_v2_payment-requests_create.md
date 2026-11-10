## lago v2 payment-requests create

Create payment request

### Synopsis

Requires a premium license. Selected customer invoices must be overdue, ready for payment processing, and share the same currency and billing entity. Payment collection and email delivery are asynchronous.
Requires the product catalog to be enabled and payment_request write permission on the API key.

```
lago v2 payment-requests create [flags]
```

### Examples

```
  lago v2 payment-requests create --input @payload.json
  lago v2 payment-requests create --input @payload.json --output json  # full resource
```

### Options

```
      --email string                                Recipient email. Defaults to the customer email.
      --external-customer-id string                 External customer identifier.
  -h, --help                                        help for create
      --input string                                Complete JSON request body or @file.json
      --lago-invoice-ids string                     Customer invoices to include.
      --payment-method-payment-method-id string     Identifier of a payment method belonging to the customer.
      --payment-method-payment-method-type string   How the payment is collected.; one of: provider, manual
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

* [lago v2 payment-requests](lago_v2_payment-requests)	 - Manage Lago payment requests
