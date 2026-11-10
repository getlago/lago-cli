## lago v2 products create

Create product

### Synopsis

Creates a fixed or metered product. Metered products require billable_metric_code; fixed products reject it.
Requires the product catalog to be enabled and product write permission on the API key.

```
lago v2 products create [flags]
```

### Examples

```
  lago v2 products create --input @payload.json
  lago v2 products create --input @payload.json --output json  # full resource
```

### Options

```
      --billable-metric-code string    Billable metric code. Required for metered products and not allowed for fixed products.
      --code string                    Unique code within the organization.
      --description string             Internal description.
  -h, --help                           help for create
      --input string                   Complete JSON request body or @file.json
      --invoice-display-name string    Name used on invoices. Defaults to the name when blank.
      --name string                    Name shown in the catalog.
      --product-category-code string   Category code. Omit or clear it to leave the product uncategorized.
      --product-type string            Type of product. Metered products require a billable metric; fixed products must not reference one.; one of: metered, fixed
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

* [lago v2 products](lago_v2_products)	 - Manage Lago products
