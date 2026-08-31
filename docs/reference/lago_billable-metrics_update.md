## lago billable-metrics update

Update a billable metric

### Synopsis

This endpoint updates an existing billable metric representing a pricing component of your application.

```
lago billable-metrics update <code> [flags]
```

### Examples

```
  lago billable-metrics update <code> --input @payload.json
```

### Options

```
      --aggregation-type string     Aggregation method used to compute usage for this billable metric.; one of: count_agg, sum_agg, max_agg, unique_count_agg, weighted_sum_agg, latest_agg
      --code string                 Unique code used to identify the billable metric associated with the API request. This code associates each event with the correct metric.
      --description string          Internal description of the billable metric.
      --expression string           Expression used to calculate the event units. The expression is evalutated for each event and the result is then used to calculate the total aggregated units.
                                    Accepted function are 'ceil', 'concat' and 'round' as well as '+', '-', '\' and '*' operations.
                                    Round is accepting an optional second parameter to specify the number of decimal.
      --field-name string           Property of the billable metric used for aggregating usage data. This field is not required for 'count_agg'.
      --filters string              API field (array)
  -h, --help                        help for update
      --idempotency-key string      Idempotency key for safe mutation retries
      --input string                Complete JSON request body or @file.json
      --name string                 Name of the billable metric.
      --recurring string            Defines if the billable metric is persisted billing period over billing period.

                                    - If set to 'true': the accumulated number of units calculated from the previous billing period is persisted to the next billing period.
                                    - If set to 'false': the accumulated number of units is reset to 0 at the end of the billing period.
                                    - If not defined in the request, default value is 'false'.
      --rounding-function string    Refers to the numeric value or mathematical expression that will be rounded based on the calculated number of billing units. Possible values are 'round', 'ceil' and 'floor'.; one of: ceil, floor, round,
      --rounding-precision string   Specifies the number of decimal places to which the 'rounding_function' will be rounded. It can be a positive or negative value.
      --weighted-interval string    Parameter exclusively utilized in conjunction with the 'weighted_sum' aggregation type. It serves to adjust the aggregation result by assigning weights and proration to the result based on time intervals. When this field is not provided, the default time interval is assumed to be in 'seconds'.; one of: seconds,
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

* [lago billable-metrics](lago_billable-metrics)	 - Manage Lago billable metrics
