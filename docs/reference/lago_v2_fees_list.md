## lago v2 fees list

List fees

### Synopsis

Returns a cursor-paginated list ordered by creation time, newest first. Use after or before to navigate. The page and per_page parameters are rejected.
Only product fees are returned. Fees can be listed before they have an invoice. Legacy fees are excluded.
Requires the product catalog to be enabled and fee read permission on the API key.

```
lago v2 fees list [flags]
```

### Examples

```
  lago v2 fees list
```

### Options

```
      --after string                  Opaque next_cursor from a previous response. Returns the next page. Cannot be combined with before or include_total_count=true. Keep the same filters when paging.
      --all                           Fetch every page by following next_cursor
      --before string                 Opaque prev_cursor from a previous response. Returns the previous page in the same list order. Cannot be combined with after or include_total_count=true. Keep the same filters when paging.
      --created-at-from string        Inclusive lower bound on the creation time, as an ISO 8601 date or timestamp. A date starts at midnight UTC. Invalid values return 422.
      --created-at-to string          Inclusive upper bound on the creation time, as an ISO 8601 date or timestamp. A date includes that whole UTC day. Invalid values return 422.
      --currency string               Currency as an ISO 4217 code.
      --event-transaction-id string   Event transaction id.
      --external-contract-id string   External contract id.
      --external-customer-id string   External customer id.
      --failed-at-from string         Inclusive lower bound on failed at. A date starts at midnight UTC.
      --failed-at-to string           Inclusive upper bound on failed at. A date-only upper bound includes the whole UTC day.
  -h, --help                          help for list
      --include-total-count string    Requests the total number of matching records. Supported only on the first page, without after or before. The count can be estimated; total_count_estimated is then present and true.
      --lago-invoice-id string        Lago invoice id.
      --limit string                  Maximum number of records per page. Values outside the supported range return 400.
      --payment-status string         Payment status.; one of: pending, succeeded, failed, refunded
      --product-code string           Product code.
      --rate-card-code string         Rate card code.
      --refunded-at-from string       Inclusive lower bound on refunded at. A date starts at midnight UTC.
      --refunded-at-to string         Inclusive upper bound on refunded at. A date-only upper bound includes the whole UTC day.
      --succeeded-at-from string      Inclusive lower bound on succeeded at. A date starts at midnight UTC.
      --succeeded-at-to string        Inclusive upper bound on succeeded at. A date-only upper bound includes the whole UTC day.
      --watch                         Poll and re-render when the response changes
      --watch-interval duration       Polling interval used with --watch (default 2s)
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

* [lago v2 fees](lago_v2_fees)	 - Manage Lago fees
