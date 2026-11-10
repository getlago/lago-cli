## lago v2 rate-cards delete

Delete rate card

### Synopsis

Deletes an unused rate card and its rates. Cards attached to plans or subscriptions cannot be deleted.
Requires the product catalog to be enabled and rate_card write permission on the API key.

```
lago v2 rate-cards delete <code> [flags]
```

### Examples

```
  lago v2 rate-cards delete <code>
  lago v2 rate-cards delete <code> --output json  # full resource
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

* [lago v2 rate-cards](lago_v2_rate-cards)	 - Manage Lago rate cards
