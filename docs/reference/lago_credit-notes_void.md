## lago credit-notes void

Void available credit

### Synopsis

This endpoint voids the available credit linked to a specific credit note.

```
lago credit-notes void <lago_id> [flags]
```

### Examples

```
  lago credit-notes void <lago_id>
  lago credit-notes void <lago_id> --output json  # full resource
```

### Options

```
  -h, --help   help for void
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

* [lago credit-notes](lago_credit-notes)	 - Manage Lago credit notes
