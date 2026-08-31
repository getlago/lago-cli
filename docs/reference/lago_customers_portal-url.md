## lago customers portal-url

Get a customer portal URL

### Synopsis

Retrieves an embeddable link for displaying a customer portal.

This endpoint allows you to fetch the URL that can be embedded to provide customers access to a dedicated portal

```
lago customers portal-url <external_customer_id> [flags]
```

### Examples

```
  lago customers portal-url <external_customer_id>
```

### Options

```
  -h, --help                      help for portal-url
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

* [lago customers](lago_customers)	 - Manage Lago customers
