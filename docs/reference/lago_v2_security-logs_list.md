## lago v2 security-logs list

List security logs

### Synopsis

Returns a cursor-paginated list ordered by log time, newest first. Use after or before to navigate. The page and per_page parameters are rejected.
Requires a premium license and available log storage. Security logs must also be enabled for the organization. Results are ordered by logged_at and log_id, both descending, within the fixed 90-day retention period.
Requires the product catalog to be enabled and security_log read permission on the API key.

```
lago v2 security-logs list [flags]
```

### Examples

```
  lago v2 security-logs list
```

### Options

```
      --after string                 Opaque next_cursor from a previous response. Returns the next page. Cannot be combined with before or include_total_count=true. Keep the same filters when paging.
      --all                          Fetch every page by following next_cursor
      --api-key-ids string           Api key ids[].
      --before string                Opaque prev_cursor from a previous response. Returns the previous page in the same list order. Cannot be combined with after or include_total_count=true. Keep the same filters when paging.
      --from-date string             Inclusive lower bound as an ISO 8601 date or timestamp. A date is interpreted as midnight UTC.
  -h, --help                         help for list
      --include-total-count string   Requests the total number of matching records. Supported only on the first page, without after or before. The count can be estimated; total_count_estimated is then present and true.
      --limit string                 Maximum number of records per page. Values outside the supported range return 400.
      --log-events string            Log events[].
      --log-types string             Log types[].
      --to-date string               Inclusive upper bound as an ISO 8601 date or timestamp. A date is interpreted as midnight UTC. Required; a request without it returns 422.
      --user-ids string              User ids[].
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

* [lago v2 security-logs](lago_v2_security-logs)	 - Manage Lago security logs
