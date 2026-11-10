## lago v2 billable-metrics create

Create billable metric

### Synopsis

Requires the product catalog to be enabled and billable_metric write permission on the API key.

```
lago v2 billable-metrics create [flags]
```

### Examples

```
  lago v2 billable-metrics create --input @payload.json
  lago v2 billable-metrics create --input @payload.json --output json  # full resource
```

### Options

```
      --aggregation-type string     Aggregation method. field_name is required except for count_agg.; one of: count_agg, sum_agg, max_agg, unique_count_agg, weighted_sum_agg, latest_agg
      --code string                 Code.
      --description string          Description.
      --expression string           Expression.
      --field-name string           Field name.
      --filters string              Complete filter definitions. An empty array removes existing filters. Null, malformed filters and non-string values are rejected before any changes are saved.
  -h, --help                        help for create
      --input string                Complete JSON request body or @file.json
      --name string                 Name.
      --recurring string            Whether values carry over between billing periods.
      --rounding-function string    Rounding applied to aggregated values.; one of: ceil, floor, round,
      --rounding-precision string   Rounding precision.
      --weighted-interval string    Weighting interval. Required for weighted_sum_agg.; one of: seconds,
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

* [lago v2 billable-metrics](lago_v2_billable-metrics)	 - Manage Lago billable metrics
