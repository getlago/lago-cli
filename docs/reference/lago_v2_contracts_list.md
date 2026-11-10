## lago v2 contracts list

List contracts

### Synopsis

Returns a cursor-paginated list ordered by creation time, newest first. Use after or before to navigate. The page and per_page parameters are rejected.

Requires the product catalog to be enabled and contract read permission on the API key.

```
lago v2 contracts list [flags]
```

### Examples

```
  lago v2 contracts list
```

### Options

```
      --after string                  Opaque next_cursor from a previous response. Returns the next page. Cannot be combined with before or include_total_count=true. Keep the same filters when paging.
      --all                           Fetch every page by following next_cursor
      --before string                 Opaque prev_cursor from a previous response. Returns the previous page in the same list order. Cannot be combined with after or include_total_count=true. Keep the same filters when paging.
      --billing-entity-ids string     Billing entity ids[].
      --created-at-from string        Inclusive lower bound on the creation time, as an ISO 8601 date or timestamp. A date starts at midnight UTC. Invalid values return 422.
      --created-at-to string          Inclusive upper bound on the creation time, as an ISO 8601 date or timestamp. A date includes that whole UTC day. Invalid values return 422.
      --external-customer-id string   External customer id.
      --external-id string            External id.
      --has-rate-overrides string     Has rate overrides.
  -h, --help                          help for list
      --include-total-count string    Requests the total number of matching records. Supported only on the first page, without after or before. The count can be estimated; total_count_estimated is then present and true.
      --limit string                  Maximum number of records per page. Values outside the supported range return 400.
      --plan-code string              Plan code.
      --search-term string            Searches contract name, external ID and Lago ID; plan name and code; and customer name, email and external ID. Ignored when external_id is provided. Customer fields are not searched when external_customer_id is provided.
      --status string                 Statuses to include. Also accepts scalar status. Defaults to active.
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

* [lago v2 contracts](lago_v2_contracts)	 - Manage Lago contracts
