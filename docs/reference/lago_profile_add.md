## lago profile add

Add or update a named profile and validate its credentials

```
lago profile add NAME [flags]
```

### Examples

```
  lago profile add staging
  lago profile add staging --api-key "$LAGO_API_KEY" --region eu --mode test --use
```

### Options

```
  -h, --help            help for add
      --region string   Lago region: us, eu, or self-hosted
      --update-check    Allow a once-daily anonymous release check
      --use             Make this profile the current profile (the first profile is current by default)
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

* [lago profile](lago_profile)	 - Manage configured Lago profiles
