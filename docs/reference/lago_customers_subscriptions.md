## lago customers subscriptions

List all customer's subscriptions

### Synopsis

This endpoint retrieves all active subscriptions for a customer.

```
lago customers subscriptions <external_customer_id> [flags]
```

### Examples

```
  lago customers subscriptions <external_customer_id>
```

### Options

```
      --all                           Fetch every page
      --billing-entity-codes string   Filter subscriptions by billing entity codes.
  -h, --help                          help for subscriptions
      --limit string                  Results per page (1-1000)
      --page string                   Page number.
      --per-page string               Number of records per page.
      --plan-code string              The unique code representing the plan to be attached to the customer. This code must correspond to the code property of one of the active plans.
      --status string                 If the field is not defined, Lago will return only 'active' subscriptions. However, if you wish to fetch subscriptions by different status you can define them in a status[] query param. Available filter values: 'pending', 'canceled', 'terminated', 'active'.
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

* [lago customers](lago_customers)	 - Manage Lago customers
