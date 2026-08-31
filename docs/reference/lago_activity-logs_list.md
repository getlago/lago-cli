## lago activity-logs list

List all activity logs

### Synopsis

This endpoint retrieves all existing activity logs that represent actions performed on application resources.

```
lago activity-logs list [flags]
```

### Examples

```
  lago activity-logs list
```

### Options

```
      --activity-sources string           Filter results by activity sources
      --activity-types string             Filter results by activity types
      --all                               Fetch every page
      --external-customer-id string       Unique identifier assigned to the customer in your application.
      --external-subscription-id string   External subscription ID
      --from-date string                  Filter activity logs from a specific date.
  -h, --help                              help for list
      --limit string                      Maximum number of results
      --page string                       Page number.
      --per-page string                   Number of records per page.
      --resource-ids string               Filter results by resources unique identifiers
      --resource-types string             Filter results by resource class types
      --to-date string                    Filter activity logs up to a specific date.
      --user-emails string                Filter results by user emails
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

* [lago activity-logs](lago_activity-logs)	 - Manage Lago activity logs
