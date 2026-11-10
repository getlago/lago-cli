## lago v2 invoices refresh

Refresh a draft invoice

### Synopsis

Refreshes a draft invoice and recalculates fees, taxes and totals. Eligible non-draft invoices return 200 unchanged. Only subscription and contract invoices support refresh; other invoice types return 403.
Requires the product catalog to be enabled and invoice write permission on the API key.

```
lago v2 invoices refresh <lago_id> [flags]
```

### Examples

```
  lago v2 invoices refresh <lago_id>
  lago v2 invoices refresh <lago_id> --output json  # full resource
```

### Options

```
  -h, --help   help for refresh
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

* [lago v2 invoices](lago_v2_invoices)	 - Manage Lago invoices
