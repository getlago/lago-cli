## lago credit-notes create

Create a credit note

### Synopsis

This endpoint creates a new credit note.

```
lago credit-notes create [flags]
```

### Examples

```
  lago credit-notes create --input @payload.json
```

### Options

```
      --credit-amount-cents string   The total amount to be credited to the customer balance for discounts on future invoices. For a total or partial refund, credit or offset, the amount in cents must include both the item amount and the applicable tax. The refunded, credited and offsetted amounts should always balance. The total, including taxes, cannot exceed the invoice's total fees.
      --description string           The description of the credit note.
  -h, --help                         help for create
      --idempotency-key string       Idempotency key for safe mutation retries
      --input string                 Complete JSON request body or @file.json
      --invoice-id string            The invoice unique identifier, created by Lago.
      --items string                 The list of credit note's items.
      --metadata string              Metadata to set as key-value pairs. Keys are strings (max 100 characters), values can be strings (max 255 characters) or null.
      --offset-amount-cents string   The total amount to be immediately deducted from the invoice balance. For a total or partial refund, credit or offset, the amount in cents must include both the item amount and the applicable tax. The refunded, credited and offsetted amounts should always balance. The total, including taxes, cannot exceed the invoice's total fees.
      --reason string                The reason of the credit note creation.
                                     Possible values are 'duplicated_charge', 'product_unsatisfactory', 'order_change', 'order_cancellation', 'fraudulent_charge' or 'other'.; one of: duplicated_charge, product_unsatisfactory, order_change, order_cancellation, fraudulent_charge, other,
      --refund-amount-cents string   The total amount to be refunded immediately to the payment method used by the customer. For a total or partial refund, credit or offset, the amount in cents must include both the item amount and the applicable tax. The refunded, credited and offsetted amounts should always balance. The total, including taxes, cannot exceed the invoice's total fees.
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

* [lago credit-notes](lago_credit-notes)	 - Manage Lago credit notes
