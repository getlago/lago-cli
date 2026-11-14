## lago usage-attribution-types list

List all usage attribution types

### Synopsis

This endpoint retrieves all the usage attribution types of the organization.

```
lago usage-attribution-types list [flags]
```

### Examples

```
  lago usage-attribution-types list
```

### Options

```
      --all                       Fetch every page
  -h, --help                      help for list
      --limit string              Results per page (1-1000)
      --page string               Page number. Defaults to 1. A value that is not a positive integer, such as '0', returns the first page.
      --per-page string           Number of records per page. Defaults to 100. The maximum is 200, unless another maximum is configured for your organization. A value above the maximum returns a '422' error with 'value_is_out_of_range' on 'per_page'. A value that is not a positive integer, such as '0', returns the default number of records.
      --role string               Filter usage attribution types by role.; one of: hierarchical, flat
      --search-term string        Filter usage attribution types by search term. The term is matched against the code and the name of the type.
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

* [lago usage-attribution-types](lago_usage-attribution-types)	 - Manage Lago usage attribution types
