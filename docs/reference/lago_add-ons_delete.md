## lago add-ons delete

Delete an add-on

### Synopsis

This endpoint is used to delete an existing add-on.

```
lago add-ons delete <code> [flags]
```

### Examples

```
  lago add-ons delete <code>
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

* [lago add-ons](lago_add-ons)	 - Manage Lago add ons
