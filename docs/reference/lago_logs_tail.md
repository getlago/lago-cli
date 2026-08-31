## lago logs tail

Continuously render API logs when they change

```
lago logs tail [flags]
```

### Examples

```
  lago logs tail
  lago logs tail --status 4xx --resource /customers --method POST
```

### Options

```
  -h, --help                help for tail
      --interval duration   Polling interval (default 2s)
      --method strings      HTTP method filter (repeatable)
      --resource string     Request path filter
      --status strings      HTTP status or class filter (repeatable)
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

* [lago logs](lago_logs)	 - Inspect Lago API request logs
