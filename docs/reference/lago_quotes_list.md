## lago quotes list

List all quotes

### Synopsis

This endpoint is used to list all existing quotes.
Quotes are ordered by creation date, from the most recent to the oldest.
This is a premium feature.

```
lago quotes list [flags]
```

### Examples

```
  lago quotes list
```

### Options

```
      --all                           Fetch every page
      --external-customer-id string   Filter quotes by the external unique identifiers of their customers (provided by your own application).
      --from-date string              Filter quotes created from a specific date.
  -h, --help                          help for list
      --limit string                  Maximum number of results
      --number string                 Filter quotes by their number, as assigned by Lago.
      --order-type string             Filter quotes by order type. Possible values are 'subscription_creation', 'subscription_amendment' and 'one_off'.
      --owner-id string               Filter quotes by the Lago identifiers of the users owning them.
      --page string                   Page number.
      --per-page string               Number of records per page.
      --status string                 Filter quotes by the status of their current version. Possible values are 'draft', 'approved' and 'voided'.
      --to-date string                Filter quotes created up to a specific date.
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

* [lago quotes](lago_quotes)	 - Manage Lago quotes
