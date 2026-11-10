## lago v2 rate-cards patch

Update rate card

### Synopsis

PUT and PATCH both accept partial updates. Omitted fields are unchanged.
Product and product filter are fixed at creation. Once any rate exists, currency, applied_pricing_unit_code, billing_timing, proration, regroup_paid_fees and display_on_invoice cannot change. Code and currency also cannot change while the card is attached to a plan or contract. Tax assignments remain editable.
Requires the product catalog to be enabled and rate_card write permission on the API key.

```
lago v2 rate-cards patch <code> [flags]
```

### Examples

```
  lago v2 rate-cards patch <code> --input @payload.json
  lago v2 rate-cards patch <code> --input @payload.json --output json  # full resource
```

### Options

```
      --applied-pricing-unit-code string   Pricing unit code, or null for currency pricing.
      --billing-timing string              When fees are billed.; one of: arrears, advance
      --code string                        Unique code within the organization.
      --currency string                    ISO 4217 billing currency.
      --description string                 Internal description.
      --display-on-invoice string          Whether generated fees are shown on invoices.
  -h, --help                               help for patch
      --input string                       Complete JSON request body or @file.json
      --name string                        Name shown in the catalog.
      --proration string                   Whether charges are prorated.
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
