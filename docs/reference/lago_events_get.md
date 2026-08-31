## lago events get

Retrieve a specific event

### Synopsis

This endpoint is used for retrieving a specific usage measurement event that has been sent to a customer or a subscription.

Note that transaction_id is unique per external_subscription_id so multiple subscriptions can share the same transaction_id. This endpoint will only return the first event found with the given transaction_id.

WARNING: If your Lago organization is configured to use the Clickhouse-based event pipeline, multiple events can share the same `transaction_id` (with different timestamps). This endpoint will only return the first event found.


```
lago events get <transaction_id> [flags]
```

### Examples

```
  lago events get <transaction_id>
```

### Options

```
  -h, --help                      help for get
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

* [lago events](lago_events)	 - Manage Lago events
