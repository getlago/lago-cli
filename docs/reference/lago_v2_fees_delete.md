## lago v2 fees delete

Delete fee

### Synopsis

Only uninvoiced fees can be deleted. A fee already attached to an invoice returns 422 invoiced_fee.
Requires the product catalog to be enabled and fee write permission on the API key.

```
lago v2 fees delete <lago_id> [flags]
```

### Examples

```
  lago v2 fees delete <lago_id>
  lago v2 fees delete <lago_id> --output json  # full resource
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

* [lago v2 fees](lago_v2_fees)	 - Manage Lago fees
