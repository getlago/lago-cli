## lago orders list

List all orders

### Synopsis

This endpoint is used to list all existing orders.
This is a premium feature.

```
lago orders list [flags]
```

### Examples

```
  lago orders list
```

### Options

```
      --all                        Fetch every page
      --customer-id string         Filter orders by the Lago identifiers of their customers.
      --executed-at-from string    Filter orders executed from a specific date and time, in UTC (ISO 8601). Orders that have not been executed are excluded.
      --executed-at-to string      Filter orders executed up to a specific date and time, in UTC (ISO 8601). Orders that have not been executed are excluded.
      --execution-mode string      Filter orders by execution mode. Possible values are 'execute_in_lago' and 'order_only'.
  -h, --help                       help for list
      --limit string               Results per page (1-1000)
      --number string              Filter orders by their number, as assigned by Lago.
      --order-form-number string   Filter orders by the number of the order form they come from.
      --order-type string          Filter orders by the order type of the quote they come from. Possible values are 'subscription_creation', 'subscription_amendment' and 'one_off'.
      --owner-id string            Filter orders by the Lago identifiers of the users owning the quote they come from.
      --page string                Page number.
      --per-page string            Number of records per page.
      --quote-number string        Filter orders by the number of the quote they come from.
      --search-term string         Search orders by number.
      --status string              Filter orders by status. Possible values are 'created', 'executed' and 'failed'.
      --watch                      Poll and re-render when the response changes
      --watch-interval duration    Polling interval used with --watch (default 2s)
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

* [lago orders](lago_orders)	 - Manage Lago orders
