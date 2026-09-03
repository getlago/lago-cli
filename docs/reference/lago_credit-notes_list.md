## lago credit-notes list

List all credit notes

### Synopsis

This endpoint list all existing credit notes.

```
lago credit-notes list [flags]
```

### Examples

```
  lago credit-notes list
```

### Options

```
      --all                            Fetch every page
      --amount-from string             Filter credit notes of at least a specific amount. This parameter must be defined in cents to ensure consistent handling for all currency types.
      --amount-to string               Filter credit notes up to a specific amount. This parameter must be defined in cents to ensure consistent handling for all currency types.
      --billing-entity-codes string    Filter credit notes by billing entity codes.
      --credit-status string           Filter credit notes by credit status. Possible values are 'available', 'consumed' or 'voided'.; one of: available, consumed, voided
      --currency string                Filter credit notes by currency. Possible values ISO 4217 currency codes.
      --external-customer-id string    Unique identifier assigned to the customer in your application.
  -h, --help                           help for list
      --invoice-number string          Filter credit notes by their related invoice number.
      --issuing-date-from string       Filter credit notes starting from a specific date.
      --issuing-date-to string         Filter credit notes up to a specific date.
      --limit string                   Results per page (1-1000)
      --page string                    Page number.
      --per-page string                Number of records per page.
      --purchase-order-number string   Filter by the invoice purchase order number. The match is exact but case-insensitive.
      --reason string                  Filter credit notes by reasons. Possible values are 'product_unsatisfactory', 'order_change', 'order_cancellation', 'fraudulent_charge', 'duplicated_charge' or 'other'.; one of: product_unsatisfactory, order_change, order_cancellation, fraudulent_charge, duplicated_charge, other
      --refund-status string           Filter credit notes by refund status. Possible values are 'pending', 'succeeded' or 'failed'.; one of: pending, succeeded, failed
      --search-term string             Search credit notes by id, number, customer name, customer external_id or customer email.
      --self-billed string             Filter credit notes belonging to a self billed invoice. Possible values are 'true' or 'false'.
      --types string                   Filter credit notes by the kind of amount they carry. 'credit' matches credit notes with a credited amount, 'refund' those with a refunded amount and 'offset' those with an amount deducted from the invoice balance. A credit note can match several types.
      --watch                          Poll and re-render when the response changes
      --watch-interval duration        Polling interval used with --watch (default 2s)
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
