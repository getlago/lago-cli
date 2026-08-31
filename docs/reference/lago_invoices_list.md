## lago invoices list

List all invoices

### Synopsis

This endpoint is used for retrieving all invoices.

```
lago invoices list [flags]
```

### Examples

```
  lago invoices list
```

### Options

```
      --all                            Fetch every page
      --amount-from string             Filter invoices of at least a specific amount. This parameter must be defined in cents to ensure consistent handling for all currency types.
      --amount-to string               Filter invoices up to a specific amount. This parameter must be defined in cents to ensure consistent handling for all currency types.
      --billing-entity-codes string    Filter invoices by billing entity codes. Possible values are the billing entity codes you have created.
      --currency string                Filter invoices by currency. Possible values ISO 4217 currency codes.
      --external-customer-id string    Unique identifier assigned to the customer in your application.
  -h, --help                           help for list
      --invoice-type string            Filter invoices by invoice type. Possible values are 'subscription', 'add_on', 'credit', 'one_off', 'advance_charges' or 'progressive_billing'.; one of: subscription, add_on, credit, one_off, advance_charges, progressive_billing
      --issuing-date-from string       Filter invoices starting from a specific date.
      --issuing-date-to string         Filter invoices up to a specific date.
      --limit string                   Maximum number of results
      --metadata[key] string           Filter invoices by metadata. Replace 'key' with the actual metadata key you want to match, and provide the corresponding value. Providing empty value will search for invoice without given metadata key. For example, 'metadata[color]=blue'.
      --page string                    Page number.
      --payment-dispute-lost string    Filter invoices with a payment dispute lost. Possible values are 'true' or 'false'.
      --payment-overdue string         Filter invoices by payment_overdue. Possible values are 'true' or 'false'.
      --payment-statuses string        Filter invoices by payment statuses. Possible values are 'pending', 'failed' or 'succeeded'.
      --per-page string                Number of records per page.
      --purchase-order-number string   Filter by the invoice purchase order number. The match is exact but case-insensitive.
      --search-term string             Search invoices by id, number, customer name, customer external_id or customer email.
      --self-billed string             Filter invoices by self billed. Possible values are 'true' or 'false'.
      --statuses string                Filter invoices by statuses. Possible values are 'draft', 'failed', 'finalized', 'pending' and 'voided'.
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

* [lago invoices](lago_invoices)	 - Manage Lago invoices
