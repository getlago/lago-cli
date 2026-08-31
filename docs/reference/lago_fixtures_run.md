## lago fixtures run

Execute a multi-step fixture with variable interpolation

```
lago fixtures run FILE [flags]
```

### Examples

```
  lago fixtures run scenario.yaml
  lago fixtures run scenario.yaml --var customer_code=example --dry-run
```

### Options

```
  -h, --help              help for run
      --var stringArray   Fixture variable as name=value (repeatable)
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

* [lago fixtures](lago_fixtures)	 - Run declarative Lago API scenarios
