## lago v2 events batch

Ingest a batch of events

### Synopsis

Accepts product-catalog events. external_subscription_id is not supported and returns 422. For ClickHouse organizations, the product event pipeline must be configured, and success acknowledges ingestion rather than completed enrichment. Newly accepted events may not appear in the list immediately.
Requires the product catalog to be enabled and event write permission on the API key.

```
lago v2 events batch [flags]
```

### Examples

```
  lago v2 events batch --input @payload.json
```

### Options

```
      --events string   Events to ingest. The default maximum is 100; deployments can configure a different maximum.
  -h, --help            help for batch
      --input string    Complete JSON request body or @file.json
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

* [lago v2 events](lago_v2_events)	 - Manage Lago events
