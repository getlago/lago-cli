## lago coupons update

Update a coupon

### Synopsis

This endpoint is used to update a coupon that can be then attached to a customer to create a discount.

```
lago coupons update <code> [flags]
```

### Examples

```
  lago coupons update <code> --input @payload.json
```

### Options

```
      --amount-cents string                       The amount of the coupon in cents. This field is required only for coupon with 'fixed_amount' type.
      --amount-currency string                    The currency of the coupon. This field is required only for coupon with 'fixed_amount' type.; one of: , AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --applies-to-billable-metric-codes string   An array of billable metric codes to which the coupon is applicable. By specifying the billable metric codes in this field, you can restrict the coupon's usage to specific metrics only.
      --applies-to-plan-codes string              An array of plan codes to which the coupon is applicable. By specifying the plan codes in this field, you can restrict the coupon's usage to specific plans only.
      --code string                               Unique code used to identify the coupon.
      --coupon-type string                        The type of the coupon. It can have two possible values: 'fixed_amount' or 'percentage'.

                                                  - If set to 'fixed_amount', the coupon represents a fixed amount discount.
                                                  - If set to 'percentage', the coupon represents a percentage-based discount.; one of: fixed_amount, percentage
      --description string                        Description of the coupon.
      --expiration string                         Specifies the type of expiration for the coupon. It can have two possible values: 'time_limit' or 'no_expiration'.

                                                  - If set to 'time_limit', the coupon has an expiration based on a specified time limit.
                                                  - If set to 'no_expiration', the coupon does not have an expiration date and remains valid indefinitely.; one of: no_expiration, time_limit
      --expiration-at string                      The expiration date and time of the coupon. This field is required only for coupons with 'expiration' set to 'time_limit'. The expiration date and time should be specified in UTC format according to the ISO 8601 datetime standard. It indicates the exact moment when the coupon will expire and is no longer valid.
      --frequency string                          The type of frequency for the coupon. It can have three possible values: 'once', 'recurring' or 'forever'.

                                                  - If set to 'once', the coupon is applicable only for a single use.
                                                  - If set to 'recurring', the coupon can be used multiple times for recurring billing periods.
                                                  - If set to 'forever', the coupon has unlimited usage and can be applied indefinitely.; one of: once, recurring, forever
      --frequency-duration string                 Specifies the number of billing periods to which the coupon applies. This field is required only for coupons with a 'recurring' frequency type
  -h, --help                                      help for update
      --idempotency-key string                    Idempotency key for safe mutation retries
      --input string                              Complete JSON request body or @file.json
      --name string                               The name of the coupon.
      --percentage-rate string                    The percentage rate of the coupon. This field is required only for coupons with a 'percentage' coupon type.
      --reusable string                           Indicates whether the coupon can be reused or not. If set to 'true', the coupon is reusable, meaning it can be applied multiple times to the same customer. If set to 'false', the coupon can only be used once and is not reusable. If not specified, this field is set to 'true' by default.
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

* [lago coupons](lago_coupons)	 - Manage Lago coupons
