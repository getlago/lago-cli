## lago wallets create

Create a wallet

### Synopsis

This endpoint is used to create a wallet with prepaid credits.

```
lago wallets create [flags]
```

### Examples

```
  lago wallets create --input @payload.json
```

### Options

```
      --applies-to-billable-metric-codes string                      An array of billable metric codes to which the wallet is applicable. By specifying the billable metric codes in this field, you can restrict the wallet's usage to specific metrics only.
      --applies-to-fee-types string                                  An array of fee types to which the wallet is applicable. By specifying the fee types in this field, you can restrict the wallet's usage to specific fee types only.
      --billing-entity-code string                                   The code of the billing entity associated with the wallet. If not provided, the customer's billing entity is used.
      --code string                                                  The code of the wallet.
      --currency string                                              The currency of the wallet.; one of: AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GHS, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --expiration-at string                                         The date and time that determines when the wallet will expire. It follows the ISO 8601 datetime format and is expressed in Coordinated Universal Time (UTC).
      --external-customer-id string                                  The customer external unique identifier (provided by your own application)
      --granted-credits string                                       The number of free granted credits. Required only if there is no paid credits.
  -h, --help                                                         help for create
      --idempotency-key string                                       Idempotency key for safe mutation retries
      --ignore-paid-top-up-limits-on-creation string                 If set to true, the wallet will ignore paid top-up limits on creation.
      --input string                                                 Complete JSON request body or @file.json
      --invoice-custom-section-invoice-custom-section-codes string   List of unique codes identifying the invoice custom sections to apply. These override the default invoice custom sections configured at the customer or billing entity level.
      --invoice-custom-section-skip-invoice-custom-sections string   Set to 'true' to exclude all invoice custom sections from PDF generation for invoices related to this resource. When 'true', 'invoice_custom_section_codes' is ignored.
      --invoice-requires-successful-payment string                   A boolean setting that, when set to true, delays issuing an invoice for a wallet top-up until a successful payment is made; if false, the invoice is issued immediately upon wallet top-up, regardless of the payment status. Default value of false.
      --metadata string                                              Metadata to set as key-value pairs. Keys are strings (max 100 characters), values can be strings (max 255 characters) or null.
      --name string                                                  The name of the wallet.
      --paid-credits string                                          The number of paid credits. Required only if there is no granted credits.
      --paid-top-up-max-amount-cents string                          Maximum amount of cents that can be top-up with a single paid transaction.
      --paid-top-up-min-amount-cents string                          Minimum amount of cents that can be top-up with a single paid transaction.
      --payment-method-payment-method-id string                      The unique identifier of the payment method (required when using a specific provider payment method).
      --payment-method-payment-method-type string                    The type of payment method to use.; one of: provider, manual
      --priority string                                              Wallet priority for ordering when a customer has multiple wallets. Allowed values: 1-50, where 1 is highest priority and 50 is lowest. Defaults to 50.
      --purchase-order-number string                                 The purchase order number associated with the wallet. It will be added to invoices generated for wallet top-ups, unless a more specific purchase order number is set on the wallet transaction or recurring transaction rule that triggered the top-up.
      --rate-amount string                                           The rate of conversion between credits and the amount in the specified currency. It indicates the ratio or factor used to convert credits into the corresponding monetary value in the currency of the transaction.
      --recurring-transaction-rules string                           List of recurring transaction rules. Currently, we only allow one recurring rule per wallet.
      --transaction-metadata string                                  This optional field allows you to store a list of key-value pairs that provide additional information or custom attributes. These key-value pairs will be included in the metadata of wallet transactions generated during the wallet creation process.
      --transaction-name string                                      The name of the wallet transactions triggered when creating the wallet. It will appear on the invoice as the label for the fee. If not set, the label on the invoice will fallback to either 'Prepaid credits - {{wallet_name}}' if the wallet name is set, or 'Prepaid credits'.
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

* [lago wallets](lago_wallets)	 - Manage Lago wallets
