## lago customers credit-notes

List all customer's credit notes

### Synopsis

This endpoint list all existing credit notes for a customer.

```
lago customers credit-notes <external_customer_id> [flags]
```

### Examples

```
  lago customers credit-notes <external_customer_id>
```

### Options

```
      --all                            Fetch every page
      --amount-from string             Filter credit notes of at least a specific amount. This parameter must be defined in cents to ensure consistent handling for all currency types.
      --amount-to string               Filter credit notes up to a specific amount. This parameter must be defined in cents to ensure consistent handling for all currency types.
      --credit-status string           Filter credit notes by credit status. Possible values are 'available', 'consumed'  or 'voided'.; one of: available, consumed, voided
  -h, --help                           help for credit-notes
      --invoice-number string          Filter credit notes by their related invoice number.
      --issuing-date-from string       Filter credit notes starting from a specific date.
      --issuing-date-to string         Filter credit notes up to a specific date.
      --limit string                   Results per page (1-1000)
      --page string                    Page number.
      --per-page string                Number of records per page.
      --purchase-order-number string   Filter by the invoice purchase order number. The match is exact but case-insensitive.
      --reason string                  Filter credit notes by reasons. Possible values are 'product_unsatisfactory', 'order_change', 'order_cancellation', 'fraudulent_charge', 'duplicated_charge' or 'other'.; one of: product_unsatisfactory, order_change, order_cancellation, fraudulent_charge, duplicated_charge, other
      --refund-status string           Filter credit notes by refund status. Possible values are 'pending', 'succeeded'  or 'failed'.; one of: pending, succeeded, failed
      --search-term string             Search credit notes by id, number, customer name, customer external_id or customer email.
      --self-billed string             Filter credit notes belonging to a self billed invoice. Possible values are 'true' or 'false'.
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

* [lago customers](lago_customers)	 - Manage Lago customers
