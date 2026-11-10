## lago v2 customers credit-notes

List customer credit notes

### Synopsis

Returns a cursor-paginated list for the customer identified in the path. No expansions are supported.
Requires the product catalog to be enabled and credit_note read permission on the API key. Unknown values for credit_status, refund_status, reason and types[] are ignored. If no supported values remain for a filter, that filter is not applied.

```
lago v2 customers credit-notes <customer_external_id> [flags]
```

### Examples

```
  lago v2 customers credit-notes <customer_external_id>
```

### Options

```
      --after string                   Opaque next_cursor from a previous response. Returns the next page. Cannot be combined with before or include_total_count=true. Keep the same filters when paging.
      --all                            Fetch every page by following next_cursor
      --amount-from string             Amount from.
      --amount-to string               Amount to.
      --before string                  Opaque prev_cursor from a previous response. Returns the previous page in the same list order. Cannot be combined with after or include_total_count=true. Keep the same filters when paging.
      --billing-entity-codes string    Billing entity codes[].
      --credit-status string           Credit status.; one of: available, consumed, voided
      --currency string                Currency as an ISO 4217 code.
  -h, --help                           help for credit-notes
      --include-total-count string     Requests the total number of matching records. Supported only on the first page, without after or before. The count can be estimated; total_count_estimated is then present and true.
      --invoice-number string          Invoice number.
      --issuing-date-from string       Issuing date from.
      --issuing-date-to string         Issuing date to.
      --limit string                   Maximum number of records per page. Values outside the supported range return 400.
      --purchase-order-number string   Purchase order number.
      --reason string                  Reason.; one of: duplicated_charge, product_unsatisfactory, order_change, order_cancellation, fraudulent_charge, other
      --refund-status string           Refund status.; one of: pending, succeeded, failed
      --search-term string             Search term.
      --self-billed string             Self billed.
      --types string                   Types[].
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

* [lago v2 customers](lago_v2_customers)	 - Manage Lago customers
