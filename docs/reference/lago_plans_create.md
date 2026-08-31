## lago plans create

Create a plan

### Synopsis

This endpoint creates a plan with subscription and usage-based charges. It supports flexible billing cadence (in-advance or in-arrears) and allows for both recurring and metered charges.

```
lago plans create [flags]
```

### Examples

```
  lago plans create --input @payload.json
```

### Options

```
      --amount-cents string                              The base cost of the plan, excluding any applicable taxes, that is billed on a recurring basis. This value is defined at 0 if your plan is a pay-as-you-go plan.
      --amount-currency string                           The currency of the plan. It indicates the monetary unit in which the plan's cost, including taxes and usage-based charges, is expressed.; one of: AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GHS, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --bill-charges-monthly string                      This field, when set to 'true', enables to invoice usage-based charges on monthly basis, even if the cadence of the plan is 'yearly' or 'semiannual'. This allows customers to pay charges overage on a monthly basis. This can be set to true only if the plan's interval is 'yearly' or 'semiannual'.
      --bill-fixed-charges-monthly string                This field, when set to 'true', enables to invoice fixed charges on monthly basis, even if the cadence of the plan is 'yearly' or 'semiannual'. This allows customers to pay fixed charges on a monthly basis. This can be set to true only if the plan's interval is 'yearly' or 'semiannual'.
      --charges string                                   Additional usage-based charges for this plan.
      --code string                                      The code of the plan. It serves as a unique identifier associated with a particular plan. The code is typically used for internal or system-level identification purposes, like assigning a subscription, for instance.
      --description string                               The description on the plan.
      --fixed-charges string                             Additional fixed charges for this plan.
  -h, --help                                             help for create
      --idempotency-key string                           Idempotency key for safe mutation retries
      --input string                                     Complete JSON request body or @file.json
      --interval string                                  The interval used for recurring billing. It represents the frequency at which subscription billing occurs. The interval can be one of the following values: 'yearly', 'semiannual', 'quarterly', 'monthly', or 'weekly'.; one of: weekly, monthly, quarterly, semiannual, yearly
      --invoice-display-name string                      Specifies the name that will be displayed on an invoice. If no value is set for this field, the name of the plan will be used as the default display name.
      --metadata string                                  Metadata to set as key-value pairs. Keys are strings (max 100 characters), values can be strings (max 255 characters) or null.
      --minimum-commitment-amount-cents string           The amount of the minimum commitment in cents.
      --minimum-commitment-invoice-display-name string   Specifies the name that will be displayed on an invoice. If no value is set for this field, the default name will be used as the display name.
      --minimum-commitment-tax-codes string              List of unique code used to identify the taxes.
      --name string                                      The name of the plan.
      --pay-in-advance string                            This field determines the billing timing for the plan. When set to 'true', the base cost of the plan is due at the beginning of each billing period. Conversely, when set to 'false', the base cost of the plan is due at the end of each billing period.
      --tax-codes string                                 List of unique code used to identify the taxes.
      --trial-period string                              The duration in days during which the base cost of the plan is offered for free.
      --usage-thresholds string                          List of usage thresholds to apply to the plan.
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

* [lago plans](lago_plans)	 - Manage Lago plans
