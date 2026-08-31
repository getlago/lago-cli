## lago invoices create

Create a one-off invoice

### Synopsis

This endpoint is used for issuing a one-off invoice.

```
lago invoices create [flags]
```

### Examples

```
  lago invoices create --input @payload.json
```

### Options

```
      --billing-entity-code string                                   The code of the billing entity to issue the invoice from. If not provided, the customer's billing entity is used.
      --currency string                                              The currency of the invoice.; one of: AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GHS, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --external-customer-id string                                  Unique identifier assigned to the customer in your application.
      --fees string                                                  API field (array)
  -h, --help                                                         help for create
      --idempotency-key string                                       Idempotency key for safe mutation retries
      --input string                                                 Complete JSON request body or @file.json
      --invoice-custom-section-invoice-custom-section-codes string   List of unique codes identifying the invoice custom sections to apply. These override the default invoice custom sections configured at the customer or billing entity level.
      --invoice-custom-section-skip-invoice-custom-sections string   Set to 'true' to exclude all invoice custom sections from PDF generation for invoices related to this resource. When 'true', 'invoice_custom_section_codes' is ignored.
      --payment-method-payment-method-id string                      The unique identifier of the payment method (required when using a specific provider payment method).
      --payment-method-payment-method-type string                    The type of payment method to use.; one of: provider, manual
      --purchase-order-number string                                 The purchase order number to associate with the invoice.
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

* [lago invoices](lago_invoices)	 - Manage Lago invoices
