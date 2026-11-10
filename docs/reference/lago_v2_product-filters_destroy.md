## lago v2 product-filters destroy

Delete filter

### Synopsis

Soft-deletes the filter, its values, its scoped rate cards and their rates. Deletion is rejected while the product is used by a plan or contract.
Requires the product catalog to be enabled and product write permission on the API key.

```
lago v2 product-filters destroy <product_code> <code> [flags]
```

### Examples

```
  lago v2 product-filters destroy <product_code> <code>
  lago v2 product-filters destroy <product_code> <code> --output json  # full resource
```

### Options

```
  -h, --help   help for destroy
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

* [lago v2 product-filters](lago_v2_product-filters)	 - Manage Lago product filters
