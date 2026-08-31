## lago subscriptions get

Retrieve a subscription

### Synopsis

This endpoint retrieves a specific subscription.

```
lago subscriptions get <external_id> [flags]
```

### Examples

```
  lago subscriptions get <external_id>
```

### Options

```
  -h, --help                      help for get
      --status string             By default, this endpoint only return 'active' subscriptions. If you want to retrieve a subscription with a different 'status', you can specify it here.

                                  _Note: As there may exists multiple 'canceled' or 'terminated' subscribtions for the same 'external_id', it is recommended to use the "List all subscriptions" endpoint to retrieve those subscriptions._; one of: active, terminated, pending, canceled
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

* [lago subscriptions](lago_subscriptions)	 - Manage Lago subscriptions
