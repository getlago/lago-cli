## lago coupons apply

Apply a coupon to a customer

### Synopsis

This endpoint is used to apply a specific coupon to a customer, before or during a subscription.

```
lago coupons apply [flags]
```

### Examples

```
  lago coupons apply --input @payload.json
```

### Options

```
      --amount-cents string           The amount of the coupon in cents. This field is required only for coupon with 'fixed_amount' type.
      --amount-currency string        The currency of the coupon. This field is required only for coupon with 'fixed_amount' type.; one of: , AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --coupon-code string            Unique code used to identify the coupon.
      --external-customer-id string   The customer external unique identifier (provided by your own application)
      --frequency string              The type of frequency for the coupon. It can have three possible values: 'once', 'recurring' or 'forever'.

                                      - If set to 'once', the coupon is applicable only for a single use.
                                      - If set to 'recurring', the coupon can be used multiple times for recurring billing periods.
                                      - If set to 'forever', the coupon has unlimited usage and can be applied indefinitely.; one of: once, recurring, forever,
      --frequency-duration string     Specifies the number of billing periods to which the coupon applies. This field is required only for coupons with a 'recurring' frequency type
  -h, --help                          help for apply
      --input string                  Complete JSON request body or @file.json
      --percentage-rate string        The percentage rate of the coupon. This field is required only for coupons with a 'percentage' coupon type.
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
