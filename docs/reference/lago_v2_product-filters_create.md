## lago v2 product-filters create

Create filter

### Synopsis

Creates a filter for a metered product only. Values must refer to keys and configured values of its billable metric. A key-only entry cannot be combined with specific values for the same key.
Requires the product catalog to be enabled and product write permission on the API key.

```
lago v2 product-filters create <product_code> [flags]
```

### Examples

```
  lago v2 product-filters create <product_code> --input @payload.json
  lago v2 product-filters create <product_code> --input @payload.json --output json  # full resource
```

### Options

```
      --code string                   Unique code within the organization.
      --description string            Internal description.
  -h, --help                          help for create
      --input string                  Complete JSON request body or @file.json
      --invoice-display-name string   Name used on invoices. Defaults to the name when blank.
      --name string                   Name shown in the catalog.
      --values string                 Matches against keys of the product billable metric. Cannot be empty.
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
