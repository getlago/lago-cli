## lago payment-requests create

Create a payment request

### Synopsis

This endpoint is used to create a payment request to collect payments of overdue invoices of a given customer

```
lago payment-requests create [flags]
```

### Examples

```
  lago payment-requests create --input @payload.json
```

### Options

```
      --email string                                The customer's email address used for sending dunning notifications
      --external-customer-id string                 The customer external unique identifier (provided by your own application)
  -h, --help                                        help for create
      --idempotency-key string                      Idempotency key for safe mutation retries
      --input string                                Complete JSON request body or @file.json
      --lago-invoice-ids string                     A list of Lago IDs for the customer's overdue invoices to start the dunning process
      --payment-method-payment-method-id string     The unique identifier of the payment method (required when using a specific provider payment method).
      --payment-method-payment-method-type string   The type of payment method to use.; one of: provider, manual
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

* [lago payment-requests](lago_payment-requests)	 - Manage Lago payment requests
