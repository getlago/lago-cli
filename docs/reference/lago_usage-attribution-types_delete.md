## lago usage-attribution-types delete

Delete a usage attribution type

### Synopsis

This endpoint deletes an existing usage attribution type, along with the attribution values recorded for it. Its children are kept.

```
lago usage-attribution-types delete <code> [flags]
```

### Examples

```
  lago usage-attribution-types delete <code>
  lago usage-attribution-types delete <code> --output json  # full resource
```

### Options

```
  -h, --help   help for delete
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

* [lago usage-attribution-types](lago_usage-attribution-types)	 - Manage Lago usage attribution types
