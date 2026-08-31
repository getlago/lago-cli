## lago subscriptions update-subscription-lifetime-usage

Update a subscription lifetime usage

### Synopsis

This endpoint allows you to update the lifetime usage of a subscription.

```
lago subscriptions update-subscription-lifetime-usage <external_id> [flags]
```

### Examples

```
  lago subscriptions update-subscription-lifetime-usage <external_id> --input @payload.json
```

### Options

```
      --external-historical-usage-amount-cents string   The historical usage amount in cents for the subscription (provided by your own application).
  -h, --help                                            help for update-subscription-lifetime-usage
      --input string                                    Complete JSON request body or @file.json
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
