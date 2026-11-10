## lago v2 rate-card-rates patch

Update rate

### Synopsis

PUT and PATCH both accept partial updates. Omitted fields are unchanged.
Terminated rates are frozen. Active rates reject effective_from, rate_model, min_amount_cents and billing interval fields, even when unchanged. Pending rates are fully editable; attachment and identity restrictions also apply.
Requires the product catalog to be enabled and rate_card write permission on the API key.

```
lago v2 rate-card-rates patch <rate_card_code> <code> [flags]
```

### Examples

```
  lago v2 rate-card-rates patch <rate_card_code> <code> --input @payload.json
  lago v2 rate-card-rates patch <rate_card_code> <code> --input @payload.json --output json  # full resource
```

### Options

```
      --applied-pricing-unit-conversion-rate string   Conversion rate for the applied pricing unit. Required when the card uses a pricing unit. Sent as a decimal string.
      --billing-interval-count string                 Number of billing interval units. Defaults to 1 on creation.
      --billing-interval-unit string                  Unit of the billing interval.; one of: day, week, month, year
      --code string                                   Code unique within the rate card.
      --effective-from string                         Effective timestamp. It cannot be before today in the organization timezone, must be unique on the card, and must be after the active rate. Arrears rates start at the beginning of the day.
  -h, --help                                          help for patch
      --input string                                  Complete JSON request body or @file.json
      --min-amount-cents string                       Minimum charge in cents. Allowed only for arrears billing. Defaults to 0 on creation.
      --rate-model string                             Pricing model. Fixed products support standard, graduated and volume. Volume cannot be billed in advance; dynamic requires a sum metric. Proration and aggregation can impose further restrictions.; one of: standard, graduated, package, percentage, volume, graduated_percentage, custom, dynamic
      --rate-properties string                        Pricing configuration for the selected rate model. Standard requires amount; package also requires package_size. Graduated, volume and graduated_percentage models use their matching ranges array. Custom properties remain model-specific.
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

* [lago v2 rate-card-rates](lago_v2_rate-card-rates)	 - Manage Lago rate card rates
