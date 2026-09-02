## lago invoices void

Void an invoice

### Synopsis

This endpoint is used for voiding an invoice.
• When no body parameters are provided, the invoice can be voided only if it is in a `finalized` status and its payment status is NOT `succeeded`.
• When `generate_credit_note` is provided (optionally with `refund_amount` and/or `credit_amount`), this validation is bypassed: the invoice is forcibly voided and a credit note is generated. If the specified refund/credit amounts do not cover the full invoice total, the remainder is issued on a second credit note that is created and immediately voided.

```
lago invoices void <lago_id> [flags]
```

### Examples

```
  lago invoices void <lago_id> --input @payload.json
```

### Options

```
      --credit-amount string          Portion of the invoice amount (in cents) to be credited to the customer's balance in the generated credit note.
      --generate-credit-note string   Set to 'true' to force voiding the invoice and generate a credit note.
  -h, --help                          help for void
      --input string                  Complete JSON request body or @file.json
      --refund-amount string          Portion of the invoice amount (in cents) to be refunded to the customer in the generated credit note.
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
