## lago v2 rate-cards create

Create rate card

### Synopsis

Requires the product catalog to be enabled and rate_card write permission on the API key.

```
lago v2 rate-cards create [flags]
```

### Examples

```
  lago v2 rate-cards create --input @payload.json
  lago v2 rate-cards create --input @payload.json --output json  # full resource
```

### Options

```
      --applied-pricing-unit-code string   Pricing unit code, or null for currency pricing.
      --billing-timing string              When fees are billed.; one of: arrears, advance
      --code string                        Unique code within the organization.
      --currency string                    ISO 4217 billing currency.
      --description string                 Internal description.
      --display-on-invoice string          Whether generated fees are shown on invoices.
  -h, --help                               help for create
      --input string                       Complete JSON request body or @file.json
      --name string                        Name shown in the catalog.
      --product-code string                Code of the product being priced.
      --product-filter-code string         Filter code belonging to this product.
      --proration string                   Whether charges are prorated.
      --rates string                       Initial effective-dated rates. A failure rolls back the card and its rates.
      --regroup-paid-fees string           Only for hidden advance fees. Null keeps paid fees standalone.; one of: invoice,
      --tax-codes string                   Tax codes to assign. An empty array clears the assignments.
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
