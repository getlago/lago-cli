## lago init

Configure a Lago profile and validate its credentials

```
lago init [flags]
```

### Examples

```
  lago init
  lago init --api-key lago_test_FAKE000000000000000000000000 --region eu --mode test
  lago init --api-key "$LAGO_API_KEY" --region self-hosted --api-url https://billing.example.test
```

### Options

```
  -h, --help            help for init
      --region string   Lago region: us, eu, or self-hosted
      --update-check    Allow a once-daily anonymous release check
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
