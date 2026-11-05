## lago subscriptions list

List all subscriptions

### Synopsis

This endpoint retrieves all active subscriptions.

```
lago subscriptions list [flags]
```

### Examples

```
  lago subscriptions list
```

### Options

```
      --all                           Fetch every page
      --billing-entity-codes string   Filter subscriptions by billing entity codes. A code that does not match a billing entity of the organization returns a '404' error.
      --currency string               Filter the results by currency, expressed as an ISO 4217 code.
      --external-customer-id string   The customer external unique identifier (provided by your own application)
      --external-id string            Filter subscriptions by their external unique identifier (provided by your own application).
  -h, --help                          help for list
      --limit string                  Results per page (1-1000)
      --overridden string             When set, filters subscriptions by whether their plan is an override of a parent plan. 'true' returns only subscriptions on overridden plans; 'false' returns only those on standard (non-overridden) plans.
      --page string                   Page number. Defaults to 1. A value that is not a positive integer, such as '0', returns the first page.
      --per-page string               Number of records per page. Defaults to 100. The maximum is 200, unless another maximum is configured for your organization. A value above the maximum returns a '422' error with 'value_is_out_of_range' on 'per_page'. A value that is not a positive integer, such as '0', returns the default number of records.
      --plan-code string              The unique code representing the plan to be attached to the customer. This code must correspond to the code property of one of the active plans.
      --status string                 If the field is not defined, Lago will return only 'active' subscriptions. However, if you wish to fetch subscriptions by different status you can define them in a status[] query param. Available filter values: 'pending', 'canceled', 'terminated', 'active', 'incomplete'.
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

* [lago subscriptions](lago_subscriptions)	 - Manage Lago subscriptions
