## lago v2 api-logs list

List api logs

### Synopsis

Returns a cursor-paginated list ordered by log time, newest first. Use after or before to navigate. The page and per_page parameters are rejected.
Requires a premium license and available log storage. Results are ordered by logged_at and request_id, both descending, within the configured retention period.
Requires the product catalog to be enabled and api_log read permission on the API key.

```
lago v2 api-logs list [flags]
```

### Examples

```
  lago v2 api-logs list
```

### Options

```
      --after string                 Opaque next_cursor from a previous response. Returns the next page. Cannot be combined with before or include_total_count=true. Keep the same filters when paging.
      --all                          Fetch every page by following next_cursor
      --api-version string           Api version.
      --before string                Opaque prev_cursor from a previous response. Returns the previous page in the same list order. Cannot be combined with after or include_total_count=true. Keep the same filters when paging.
      --clients string               Clients[].
      --from-date string             Inclusive lower log-time bound as an ISO 8601 date or timestamp. A date is interpreted as midnight UTC.
  -h, --help                         help for list
      --http-methods string          Http methods[].
      --http-statuses string         HTTP status codes or status groups. succeeded selects codes up to 399; failed selects codes above 399. Groups take precedence over numeric codes. Selecting both groups returns no matches.
      --include-total-count string   Requests the total number of matching records. Supported only on the first page, without after or before. The count can be estimated; total_count_estimated is then present and true.
      --limit string                 Maximum number of records per page. Values outside the supported range return 400.
      --request-paths string         Request paths[].
      --to-date string               Inclusive upper log-time bound as an ISO 8601 date or timestamp. A date is interpreted as midnight UTC.
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

* [lago v2 api-logs](lago_v2_api-logs)	 - Manage Lago api logs
