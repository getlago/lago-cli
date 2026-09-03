## lago add-ons create

Create an add-on

### Synopsis

This endpoint is used to create an add-on that can be then attached to a one-off invoice.

```
lago add-ons create [flags]
```

### Examples

```
  lago add-ons create --input @payload.json
  lago add-ons create --input @payload.json --output json  # full resource
```

### Options

```
      --amount-cents string           The cost of the add-on in cents, excluding any applicable taxes, that is billed to a customer. By creating a one-off invoice, you will be able to override this value.
      --amount-currency string        The currency of the add-on.; one of: AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GHS, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --code string                   Unique code used to identify the add-on.
      --description string            The description of the add-on.
  -h, --help                          help for create
      --input string                  Complete JSON request body or @file.json
      --invoice-display-name string   Specifies the name that will be displayed on an invoice. If no value is set for this field, the name of the actual charge will be used as the default display name.
      --name string                   The name of the add-on.
      --tax-codes string              List of unique code used to identify the taxes.
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

* [lago add-ons](lago_add-ons)	 - Manage Lago add ons
