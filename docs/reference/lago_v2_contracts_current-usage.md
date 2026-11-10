## lago v2 contracts current-usage

Retrieve contract current usage

### Synopsis

Returns usage across the open billing windows of the active or pending contract. Pending contracts have no usage. Totals may span multiple invoices.
Requires the product catalog to be enabled and customer_usage read permission on the API key.

```
lago v2 contracts current-usage <external_id> [flags]
```

### Examples

```
  lago v2 contracts current-usage <external_id>
```

### Options

```
      --apply-taxes string           Apply taxes.
  -h, --help                         help for current-usage
      --product-code string          Product code.
      --product-filter-code string   Product filter code.
      --product-type string          Product type.; one of: metered, fixed
      --rate-card-code string        Rate card code.
      --watch                        Poll and re-render when the response changes
      --watch-interval duration      Polling interval used with --watch (default 2s)
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

* [lago v2 contracts](lago_v2_contracts)	 - Manage Lago contracts
