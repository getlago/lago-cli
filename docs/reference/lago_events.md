## lago events

Manage Lago events

### Options

```
  -h, --help   help for events
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

* [lago](lago)	 - The official CLI for Lago billing
* [lago events batch](lago_events_batch)	 - Batch multiple events
* [lago events batch-estimate-instant-fees](lago_events_batch-estimate-instant-fees)	 - Batch estimate instant fees for a pay in advance charge
* [lago events estimate-fees](lago_events_estimate-fees)	 - Estimate fees for a pay in advance charge
* [lago events estimate-instant-fees](lago_events_estimate-instant-fees)	 - Estimate instant fees for a pay in advance charge
* [lago events get](lago_events_get)	 - Retrieve a specific event
* [lago events list](lago_events_list)	 - List all events
* [lago events send](lago_events_send)	 - Send usage events
