## lago v2 events send

Ingest an event

### Synopsis

Accepts product-catalog events. external_subscription_id is not supported and returns 422. For ClickHouse organizations, the product event pipeline must be configured, and success acknowledges ingestion rather than completed enrichment. Newly accepted events may not appear in the list immediately.
Requires the product catalog to be enabled and event write permission on the API key.

```
lago v2 events send [flags]
```

### Examples

```
  lago v2 events send --input @payload.json
```

### Options

```
      --code string                         Event code.
      --concurrency int                     Concurrent bulk event requests (1-64) (default 4)
      --external-contract-id string         External contract identifier.
      --file string                         Stream newline-delimited JSON events from a file, or - for stdin
  -h, --help                                help for send
      --input string                        Complete JSON request body or @file.json
      --precise-total-amount-cents string   Finite amount in cents, as a number or decimal string.
      --properties string                   Event properties. Defaults to an empty object.
      --timestamp string                    Unix timestamp in seconds. Defaults to the request time. Must be on or after 1970-01-01 and before 2300-01-01; deployment-specific time limits may also apply.
      --transaction-id string               Transaction identifier supplied by the sender.
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
