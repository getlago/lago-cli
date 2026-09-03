## lago events batch-estimate-instant-fees

Batch estimate instant fees for a pay in advance charge

### Synopsis

Estimate the fees that would be created after reception of an event for a billable metric attached to one or multiple pay in advance standard or percentage charges

```
lago events batch-estimate-instant-fees [flags]
```

### Examples

```
  lago events batch-estimate-instant-fees --input @payload.json
```

### Options

```
      --events string   API field (array)
  -h, --help            help for batch-estimate-instant-fees
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

* [lago events](lago_events)	 - Manage Lago events
