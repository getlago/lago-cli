## lago seed demo

Create a metric, plan, customer, subscription, event, and invoice preview

```
lago seed demo [flags]
```

### Examples

```
  lago seed demo
  lago seed demo --prefix local-demo --output json
```

### Options

```
  -h, --help            help for demo
      --prefix string   Unique external-ID prefix
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

* [lago seed](lago_seed)	 - Populate a test account with reproducible data
