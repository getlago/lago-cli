## lago invoices update

Update an invoice

### Synopsis

This endpoint is used for updating an existing invoice. A voided invoice cannot be updated and returns a `405` error.

```
lago invoices update <lago_id> [flags]
```

### Examples

```
  lago invoices update <lago_id> --input @payload.json
  lago invoices update <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help                     help for update
      --idempotency-key string   Idempotency key for safe mutation retries
      --input string             Complete JSON request body or @file.json
      --metadata string          API field (array)
      --payment-status string    The payment status of the invoice. Possible values are 'pending', 'failed' or 'succeeded'.; one of: pending, succeeded, failed
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

* [lago invoices](lago_invoices)	 - Manage Lago invoices
