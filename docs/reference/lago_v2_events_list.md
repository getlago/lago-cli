## lago v2 events list

List events

### Synopsis

Lists events using event-time filters. ClickHouse returns enriched events ordered by timestamp, external_contract_id, code and transaction_id, all descending. PostgreSQL orders by creation time and identifier, both descending. Cursors are opaque and must not be constructed from response fields. external_subscription_id, created_at_from and created_at_to are rejected. No expansions are supported.
Requires the product catalog to be enabled and event read permission on the API key.

```
lago v2 events list [flags]
```

### Examples

```
  lago v2 events list
```

### Options

```
      --after string                       Opaque next_cursor from a previous response. Returns the next page. Cannot be combined with before or include_total_count=true. Keep the same filters when paging.
      --all                                Fetch every page by following next_cursor
      --before string                      Opaque prev_cursor from a previous response. Returns the previous page in the same list order. Cannot be combined with after or include_total_count=true. Keep the same filters when paging.
      --code string                        Code.
      --external-contract-id string        External contract id.
  -h, --help                               help for list
      --include-total-count string         Requests the total number of matching records. Supported only on the first page, without after or before. The count can be estimated; total_count_estimated is then present and true.
      --limit string                       Maximum number of records per page. Values outside the supported range return 400.
      --timestamp-from string              Inclusive lower bound as an ISO 8601 date or timestamp. A date is interpreted as midnight UTC.
      --timestamp-from-started-at string   Timestamp from started at.
      --timestamp-to string                Inclusive upper bound as an ISO 8601 date or timestamp. A date is interpreted as midnight UTC.
      --transaction-id string              Transaction id.
      --watch                              Poll and re-render when the response changes
      --watch-interval duration            Polling interval used with --watch (default 2s)
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

* [lago v2 events](lago_v2_events)	 - Manage Lago events
