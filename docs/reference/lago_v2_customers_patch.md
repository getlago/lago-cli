## lago v2 customers patch

Update customer

### Synopsis

PUT and PATCH both accept partial updates. Omitted fields are unchanged. When metadata or integration_customers is supplied, it replaces the entire corresponding list. Send each existing entry with its id to keep or update it; omitted entries are deleted. Use the response lago_id as the request id. An empty array removes all entries.
The path identifies the existing customer. A missing customer returns 404. Billing and payment settings are top-level fields. Relationships are not expanded in the update response.
Requires the product catalog to be enabled and customer write permission on the API key.

```
lago v2 customers patch <external_id> [flags]
```

### Examples

```
  lago v2 customers patch <external_id> --input @payload.json
  lago v2 customers patch <external_id> --input @payload.json --output json  # full resource
```

### Options

```
      --account-type string                               Account type. Requires revenue share to be enabled. On update, changes are ignored if the customer has subscriptions, invoices, add-ons, wallets or currency-based coupons.; one of: customer, partner
      --address-line1 string                              Address line1.
      --address-line2 string                              Address line2.
      --billing-entity-code string                        Code of the customer billing entity.
      --city string                                       City.
      --country string                                    Country.
      --currency string                                   Currency as an ISO 4217 code.
      --customer-type string                              Type of customer.; one of: company, individual,
      --document-locale string                            Document locale.
      --email string                                      Email.
      --external-id string                                Unique customer identifier in your application. On update, changes are ignored if the customer has subscriptions, invoices, add-ons, wallets or currency-based coupons.
      --external-salesforce-id string                     External salesforce id.
      --finalize-zero-amount-invoice string               Whether to finalize zero-amount invoices. Null resets this setting to inherit.; one of: inherit, skip, finalize,
      --firstname string                                  Firstname.
  -h, --help                                              help for patch
      --input string                                      Complete JSON request body or @file.json
      --integration-customers string                      Integration mappings. Only one entry per integration type is accepted. On update, this is a full replacement. Include id for each existing mapping to keep or update it. Omitted mappings are deleted.
      --invoice-custom-section-codes string               Invoice custom section codes to assign. An empty array removes the assignments.
      --invoice-grace-period string                       Invoice grace period. Requires a premium license; otherwise the value is ignored.
      --lastname string                                   Lastname.
      --legal-name string                                 Legal name.
      --legal-number string                               Legal number.
      --logo-url string                                   Logo url.
      --metadata string                                   Customer metadata. At most five entries. On update, this is a full replacement. Include id for each existing entry to keep or update it. Entries without id are created, not matched by key. Omitted entries are deleted.
      --name string                                       Name.
      --payment-provider string                           Payment provider.
      --payment-provider-code string                      Payment provider code.
      --periodic-invoice-issuing-date-adjustment string   Adjustment applied to periodic invoice issuing dates.; one of: keep_anchor, align_with_finalization_date,
      --periodic-invoice-issuing-date-anchor string       Date used to issue periodic invoices.; one of: current_period_end, next_period_start,
      --phone string                                      Phone.
      --provider-customer-id string                       Customer identifier in the payment provider.
      --provider-payment-methods string                   Payment methods allowed by the provider.
      --shipping-address-address-line1 string             Address line1.
      --shipping-address-address-line2 string             Address line2.
      --shipping-address-city string                      City.
      --shipping-address-country string                   Country.
      --shipping-address-state string                     State.
      --shipping-address-zipcode string                   Zipcode.
      --skip-invoice-custom-sections string               Skip invoice custom sections.
      --state string                                      State.
      --sync string                                       Whether to run payment-provider synchronization before returning the response rather than in a background job.
      --sync-with-provider string                         Whether to create or synchronize the customer with the selected payment provider.
      --tax-codes string                                  Tax codes to assign. An empty array removes the assignments.
      --tax-identification-number string                  Tax identification number.
      --timezone string                                   Timezone. Requires a premium license; otherwise the value is ignored.
      --url string                                        Url.
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

* [lago v2 customers](lago_v2_customers)	 - Manage Lago customers
