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
      --limit string                      Maximum number of results
      --page string                       Page number.
      --per-page string                   Number of records per page.
      --periods-count string              Number of past billing period to returns in the result
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
