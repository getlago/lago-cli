## lago webhook-endpoints update

Update a webhook endpoint

### Synopsis

This endpoint is used to update an existing webhook endpoint.

```
lago webhook-endpoints update <lago_id> [flags]
```

### Examples

```
  lago webhook-endpoints update <lago_id> --input @payload.json
  lago webhook-endpoints update <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
      --event-types string      A list of event types that will trigger the webhook. Passing null means that all event types will be sent.
  -h, --help                    help for update
      --input string            Complete JSON request body or @file.json
      --name string             The name of the webhook.
      --signature-algo string   The signature used for the webhook. If no value is passed,; one of: jwt, hmac,
      --webhook-url string      The URL of the webhook endpoint.
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

* [lago webhook-endpoints](lago_webhook-endpoints)	 - Manage Lago webhook endpoints
