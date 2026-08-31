## lago fees list

List all fees

### Synopsis

This endpoint is used for retrieving all fees that has been issued.

```
lago fees list [flags]
```

### Examples

```
  lago fees list
```

### Options

```
      --all                               Fetch every page
      --billable-metric-code string       Filter results by the 'code' of the billable metric attached to the fee. Only applies to 'charge' types.
      --created-at-from string            Filter results created after creation date and time in UTC.
      --created-at-to string              Filter results created before creation date and time in UTC.
      --currency string                   Filter results by fee"s currency.; one of: AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GHS, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --event-transaction-id string       Filter results by event transaction ID
      --external-customer-id string       Unique identifier assigned to the customer in your application.
      --external-subscription-id string   External subscription ID
      --failed-at-from string             Filter results with payment failure after creation date and time in UTC.
      --failed-at-to string               Filter results with payment failure after creation date and time in UTC.
      --fee-type string                   The fee type. Possible values are 'add-on', 'charge', 'credit', 'subscription' or "commitment".; one of: charge, add_on, subscription, credit, commitment
  -h, --help                              help for list
      --limit string                      Maximum number of results
      --page string                       Page number.
      --payment-status string             Indicates the payment status of the fee. It represents the current status of the payment associated with the fee. The possible values for this field are 'pending', 'succeeded', 'failed' and refunded'.; one of: pending, succeeded, failed, refunded
      --per-page string                   Number of records per page.
      --refunded-at-from string           Filter results with payment refund after creation date and time in UTC.
      --refunded-at-to string             Filter results with payment refund after creation date and time in UTC.
      --succeeded-at-from string          Filter results with payment success after creation date and time in UTC.
      --succeeded-at-to string            Filter results with payment success after creation date and time in UTC.
      --watch                             Poll and re-render when the response changes
      --watch-interval duration           Polling interval used with --watch (default 2s)
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

* [lago fees](lago_fees)	 - Manage Lago fees
