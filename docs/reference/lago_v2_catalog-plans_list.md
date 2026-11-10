## lago v2 catalog-plans list

List catalog plans

### Synopsis

Returns a cursor-paginated list ordered by creation time, newest first. Use after or before to navigate. The page and per_page parameters are rejected.

Requires the product catalog to be enabled and plan read permission on the API key.

```
lago v2 catalog-plans list [flags]
```

### Examples

```
  lago v2 catalog-plans list
```

### Options

```
      --after string                 Opaque next_cursor from a previous response. Returns the next page. Cannot be combined with before or include_total_count=true. Keep the same filters when paging.
      --all                          Fetch every page by following next_cursor
      --before string                Opaque prev_cursor from a previous response. Returns the previous page in the same list order. Cannot be combined with after or include_total_count=true. Keep the same filters when paging.
      --created-at-from string       Inclusive lower bound on the creation time, as an ISO 8601 date or timestamp. A date starts at midnight UTC. Invalid values return 422.
      --created-at-to string         Inclusive upper bound on the creation time, as an ISO 8601 date or timestamp. A date includes that whole UTC day. Invalid values return 422.
  -h, --help                         help for list
      --include-total-count string   Requests the total number of matching records. Supported only on the first page, without after or before. The count can be estimated; total_count_estimated is then present and true.
      --limit string                 Maximum number of records per page. Values outside the supported range return 400.
      --product-category-id string   Product category id[].
      --product-filter-id string     Product filter id[].
      --product-id string            Product id[].
      --rate-card-id string          Rate card id[].
      --search-term string           Searches names and codes.
      --watch                        Poll and re-render when the response changes
      --watch-interval duration      Polling interval used with --watch (default 2s)
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

* [lago v2 catalog-plans](lago_v2_catalog-plans)	 - Manage Lago catalog plans
