## lago v2 billing-entities patch

Update billing entity

### Synopsis

PUT and PATCH both accept partial updates. Omitted fields are unchanged.
The code cannot be changed. Billing configuration fields are top-level fields. Empty tax_codes or invoice_custom_section_codes arrays remove the corresponding assignments.
Requires the product catalog to be enabled and billing_entity write permission on the API key.

```
lago v2 billing-entities patch <code> [flags]
```

### Examples

```
  lago v2 billing-entities patch <code> --input @payload.json
  lago v2 billing-entities patch <code> --input @payload.json --output json  # full resource
```

### Options

```
      --address-line1 string                              Address line1.
      --address-line2 string                              Address line2.
      --city string                                       City.
      --code string                                       Unique code among active billing entities.
      --country string                                    Country.
      --default-currency string                           Default currency.
      --document-locale string                            Document locale.
      --document-number-prefix string                     Document number prefix.
      --document-numbering string                         Document numbering scope.; one of: per_customer, per_billing_entity
      --einvoicing string                                 Einvoicing.
      --email string                                      Email.
      --email-settings string                             Enabled email notification types.
      --eu-tax-management string                          Eu tax management.
      --finalize-zero-amount-invoice string               Finalize zero amount invoice.
  -h, --help                                              help for patch
      --input string                                      Complete JSON request body or @file.json
      --invoice-custom-section-codes string               Invoice custom section codes to assign. An empty array removes the assignments.
      --invoice-footer string                             Invoice footer.
      --invoice-grace-period string                       Invoice grace period.
      --legal-name string                                 Legal name.
      --legal-number string                               Legal number.
      --logo string                                       PNG or JPEG logo encoded as a base64 data URL.
      --name string                                       Name.
      --periodic-invoice-issuing-date-adjustment string   Adjustment applied to periodic invoice issuing dates.; one of: keep_anchor, align_with_finalization_date
      --periodic-invoice-issuing-date-anchor string       Date used to issue periodic invoices.; one of: current_period_end, next_period_start
      --phone string                                      Phone.
      --state string                                      State.
      --tax-codes string                                  Tax codes to assign. An empty array removes the assignments.
      --tax-identification-number string                  Tax identification number.
      --timezone string                                   Timezone.
      --zipcode string                                    Zipcode.
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

* [lago v2 billing-entities](lago_v2_billing-entities)	 - Manage Lago billing entities
