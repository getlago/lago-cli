## lago order-forms list

List all order forms

### Synopsis

This endpoint is used to list all existing order forms.
This is a premium feature.

```
lago order-forms list [flags]
```

### Examples

```
  lago order-forms list
```

### Options

```
      --all                       Fetch every page
      --created-at-from string    Filter order forms created from a specific date and time, in UTC (ISO 8601).
      --created-at-to string      Filter order forms created up to a specific date and time, in UTC (ISO 8601).
      --customer-id string        Filter order forms by the Lago identifiers of their customers.
      --expires-at-from string    Filter order forms expiring from a specific date and time, in UTC (ISO 8601). Order forms that never expire are excluded.
      --expires-at-to string      Filter order forms expiring up to a specific date and time, in UTC (ISO 8601). Order forms that never expire are excluded.
  -h, --help                      help for list
      --limit string              Results per page (1-1000)
      --number string             Filter order forms by their number, as assigned by Lago.
      --owner-id string           Filter order forms by the Lago identifiers of the users owning the quote they come from.
      --page string               Page number.
      --per-page string           Number of records per page.
      --quote-number string       Filter order forms by the number of the quote they come from.
      --search-term string        Search order forms by number.
      --status string             Filter order forms by status. Possible values are 'generated', 'signed', 'expired' and 'voided'.
      --watch                     Poll and re-render when the response changes
      --watch-interval duration   Polling interval used with --watch (default 2s)
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

* [lago order-forms](lago_order-forms)	 - Manage Lago order forms
