## lago v2 webhook-endpoints patch

Update webhook endpoint

### Synopsis

PUT and PATCH both accept partial updates. Omitted fields are unchanged. Switching to standard_webhooks creates the endpoint's signing key, or reuses the key kept from an earlier switch; request expands[]=signing_keys to read it.

Requires the product catalog to be enabled and webhook_endpoint write permission on the API key.

```
lago v2 webhook-endpoints patch <lago_id> [flags]
```

### Examples

```
  lago v2 webhook-endpoints patch <lago_id> --input @payload.json
  lago v2 webhook-endpoints patch <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
      --event-types string      Event types to deliver. Null or ["*"] selects all events. Names are trimmed, lowercased and deduplicated. An empty array selects none.
      --expands string          Relationships to include in this response. Only one expansion level is supported.
  -h, --help                    help for patch
      --input string            Complete JSON request body or @file.json
      --name string             Endpoint name.
      --signature-algo string   Webhook signature algorithm. Only 'standard_webhooks' is accepted, and it is the default on create. Another value returns 422 with 'value_is_invalid' on 'signature_algo' and the accepted algorithms in 'allowed_values'. An existing endpoint keeps its algorithm when the field is omitted or sent with its current value.; one of: standard_webhooks
      --webhook-url string      Public URL receiving webhook requests. Private addresses are rejected.
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

* [lago v2 webhook-endpoints](lago_v2_webhook-endpoints)	 - Manage Lago webhook endpoints
