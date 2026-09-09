## lago customers payments

List all customer's payments

### Synopsis

List payments with filters combined using AND and values within each array combined using OR. Results keep their existing visibility rules and are ordered by creation date descending, then ID. Amounts are integer cents and support the full signed 64-bit range from 0 through 9223372036854775807. Unknown enum or currency values, negative or out-of-range amounts, inverted amount bounds, invalid invoice UUIDs and receipt or invoice numbers longer than 255 characters return 422 validation_errors. Invalid date bounds are ignored. Search narrows the scope before exact filters apply; invoice search is skipped when invoice_id or invoice_number is supplied, and customer search is skipped when external_customer_id is supplied. Repeat the same filters when requesting the page number returned in meta.next_page.

```
lago customers payments <external_customer_id> [flags]
```

### Examples

```
  lago customers payments <external_customer_id>
```

### Options

```
      --all                            Fetch every page
      --amount-from string             Inclusive minimum payment amount in integer cents, from 0 through 9223372036854775807; it must not exceed 'amount_to' when both bounds are supplied.
      --amount-to string               Inclusive maximum payment amount in integer cents, from 0 through 9223372036854775807; set it equal to 'amount_from' to match an exact amount.
      --created-at-from string         Filter payments created on or after this ISO-8601 date, inclusive from the start of the day in the organization timezone; invalid dates are ignored.
      --created-at-to string           Filter payments created on or before this ISO-8601 date, inclusive through the end of the day in the organization timezone; invalid dates are ignored.
      --currency string                Filter the results by currency, expressed as an ISO 4217 code.
  -h, --help                           help for payments
      --invoice-id string              Filter by the Lago invoice UUID, matching the directly payable invoice or any invoice covered by a payment request.
      --invoice-number string          Filter by an exact, case-insensitive invoice number of at most 255 characters, matching either the directly payable invoice or any invoice covered by a payment request.
      --limit string                   Results per page (1-1000)
      --page string                    Page number.
      --payable-type string            Filter by either 'Invoice' or 'PaymentRequest', matching any supplied payable type; a single value can also be sent as 'payable_type=PaymentRequest'.
      --payment-provider-type string   Filter by any of 'stripe', 'gocardless', 'cashfree', 'adyen', 'flutterwave' or 'moneyhash'; a single value can also be sent as 'payment_provider_type=stripe'.
      --payment-status string          Filter by any of 'pending', 'processing', 'succeeded' or 'failed'; a single value can also be sent as 'payment_status=succeeded', and this parameter takes precedence over 'payment_statuses'.
      --payment-statuses string        Alias for 'payment_status[]', matching any of 'pending', 'processing', 'succeeded' or 'failed'; a single value can also be sent as 'payment_statuses=succeeded', and it is ignored when 'payment_status' is supplied.
      --payment-type string            Filter by either 'manual' or 'provider', matching any supplied type; a single value can also be sent as 'payment_type=manual'.
      --per-page string                Number of records per page.
      --receipt-number string          Filter by an exact, case-insensitive payment receipt number of at most 255 characters; payments without a receipt do not match.
      --search-term string             Search case-insensitively within provider payment IDs, references, payment UUIDs, directly payable invoice numbers and customer name, first name, last name, external ID or email; receipt numbers use their own exact filter.
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
