## lago customers projected-usage

Retrieve customer current and projected usage

### Synopsis

This endpoint enables the retrieval of the usage-based billing data for a customer within the current period. It also returns the projected usage for the current period based on the current usage.

```
lago customers projected-usage <external_customer_id> [flags]
```

### Examples

```
  lago customers projected-usage <external_customer_id>
```

### Options

```
      --apply-taxes string                Optional flag to determine if taxes should be applied. Defaults to 'true' if not provided or if null.
      --billable-metric-code string       Filter usage to a specific billable metric by its code.
      --charge-code string                Filter usage to a specific charge by its code.
      --charge-id string                  Filter usage to a specific charge by its Lago ID (UUID).
      --external-subscription-id string   The unique identifier of the subscription within your application.
      --filter-by-presentation string     Filter 'presentation_breakdowns' by a JSON-encoded array of presentation group key values. Only breakdowns matching the provided values will be returned. Pass an empty array to disable 'presentation_breakdowns' entirely.
      --group string                      Filter usage by pricing group. Pass key/value pairs as query parameters, e.g. 'group[cloud]=aws'.
  -h, --help                              help for projected-usage
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
