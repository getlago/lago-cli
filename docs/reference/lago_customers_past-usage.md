## lago customers past-usage

Retrieve customer past usage

### Synopsis

This endpoint enables the retrieval of the usage-based billing data for a customer within past periods.

```
lago customers past-usage <external_customer_id> [flags]
```

### Examples

```
  lago customers past-usage <external_customer_id>
```

### Options

```
      --all                               Fetch every page
      --billable-metric-code string       Billable metric code filter to apply to the charge usage
      --external-subscription-id string   The unique identifier of the subscription within your application.
  -h, --help                              help for past-usage
      --limit string                      Results per page (1-1000)
      --page string                       Page number. Defaults to 1. A value that is not a positive integer, such as '0', returns the first page.
      --per-page string                   Number of records per page. Defaults to 100. The maximum is 200, unless another maximum is configured for your organization. A value above the maximum returns a '422' error with 'value_is_out_of_range' on 'per_page'. A value that is not a positive integer, such as '0', returns the default number of records.
      --periods-count string              Number of past billing periods to return in the result.
      --watch                             Poll and re-render when the response changes
      --watch-interval duration           Polling interval used with --watch (default 2s)
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
