## lago webhooks json-public-key

Retrieve webhook public key (JSON)

### Synopsis

This endpoint is used to retrieve the public key used to verify the webhooks signature, wrapped in a JSON object. Prefer this endpoint over the deprecated GET /webhooks/public_key.

```
lago webhooks json-public-key [flags]
```

### Examples

```
  lago webhooks json-public-key
```

### Options

```
  -h, --help                      help for json-public-key
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

* [lago webhooks](lago_webhooks)	 - Manage Lago webhooks
