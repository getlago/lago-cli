## lago customers create

Create a customer

### Synopsis

This endpoint creates a new customer.

```
lago customers create [flags]
```

### Examples

```
  lago customers create --input @payload.json
```

### Options

```
      --account-type string                                                         The type of the account. It can have one of the following values:
                                                                                    - 'customer': the account is a customer, default value.
                                                                                    - 'partner': the account is a partner.

                                                                                    This field is only applied when revenue share is enabled on your organization; otherwise it is ignored.; one of: customer, partner
      --address-line1 string                                                        The first line of the billing address
      --address-line2 string                                                        The second line of the billing address
      --billing-configuration-document-locale string                                The document locale, specified in the ISO 639-1 format. This field represents the language or locale used for the documents issued by Lago
      --billing-configuration-invoice-grace-period string                           The grace period, expressed in days, for the invoice. This period refers to the additional time granted to the customer beyond the invoice due date to adjust usage and line items
      --billing-configuration-payment-provider string                               The payment provider utilized to initiate payments for invoices issued by Lago.
                                                                                    Accepted values: 'stripe', 'adyen', 'cashfree', 'flutterwave', 'gocardless', 'moneyhash' or null. This field is required if you intend to assign a 'provider_customer_id'.; one of: stripe, adyen, cashfree, flutterwave, gocardless, moneyhash
      --billing-configuration-payment-provider-code string                          Unique code used to identify a payment provider connection.
      --billing-configuration-provider-customer-id string                           The customer ID within the payment provider's system. If this field is not provided, Lago has the option to create a new customer record within the payment provider's system on behalf of the customer
      --billing-configuration-provider-payment-methods string                       Specifies the available payment methods that can be used for this customer when 'payment_provider' is set to 'stripe'. The 'provider_payment_methods' field is an array that allows multiple payment options to be defined. If this field is not explicitly set, the payment methods will be set to 'card'. For now, possible values are 'card', 'sepa_debit', 'us_bank_account', 'bacs_debit', 'boleto', 'link', 'crypto' and 'customer_balance'. Note that when 'link' is selected, 'card' should also be provided in the array. When 'customer_balance' is selected, no other payment can be selected.
      --billing-configuration-subscription-invoice-issuing-date-adjustment string   The logic applied on top of the subscription_invoice_issuing_date_anchor rule. You can opt to use the invoice finalization date, that includes any configured grace period.; one of: align_with_finalization_date, keep_anchor,
      --billing-configuration-subscription-invoice-issuing-date-anchor string       Defines whether the issuing date follows the current billing period's end date or the next period starting date.; one of: current_period_end, next_period_start,
      --billing-configuration-sync string                                           Set this field to 'true' if you want to create the customer in the payment provider synchronously with the customer creation process in Lago. This option is applicable only when the 'provider_customer_id' is 'null' and the customer is automatically created in the payment provider through Lago. By default, the value is set to 'false'
      --billing-configuration-sync-with-provider string                             Set this field to 'true' if you want to create a customer record in the payment provider's system. This option is applicable only when the 'provider_customer_id' is null and the 'sync_with_provider' field is set to 'true'. By default, the value is set to 'false'
      --billing-entity-code string                                                  The unique code of the billing entity to associate with the customer. If not provided, the default billing entity will be used.
      --city string                                                                 The city of the customer's billing address
      --country string                                                              Country code of the customer's billing address. Format must be ISO 3166 (alpha-2); one of: , AD, AE, AF, AG, AI, AL, AM, AO, AQ, AR, AS, AT, AU, AW, AX, AZ, BA, BB, BD, BE, BF, BG, BH, BI, BJ, BL, BM, BN, BO, BQ, BR, BS, BT, BV, BW, BY, BZ, CA, CC, CD, CF, CG, CH, CI, CK, CL, CM, CN, CO, CR, CU, CV, CW, CX, CY, CZ, DE, DJ, DK, DM, DO, DZ, EC, EE, EG, EH, ER, ES, ET, FI, FJ, FK, FM, FO, FR, GA, GB, GD, GE, GF, GG, GH, GI, GL, GM, GN, GP, GQ, GR, GS, GT, GU, GW, GY, HK, HM, HN, HR, HT, HU, ID, IE, IL, IM, IN, IO, IQ, IR, IS, IT, JE, JM, JO, JP, KE, KG, KH, KI, KM, KN, KP, KR, KW, KY, KZ, LA, LB, LC, LI, LK, LR, LS, LT, LU, LV, LY, MA, MC, MD, ME, MF, MG, MH, MK, ML, MM, MN, MO, MP, MQ, MR, MS, MT, MU, MV, MW, MX, MY, MZ, NA, NC, NE, NF, NG, NI, NL, NO, NP, NR, NU, NZ, OM, PA, PE, PF, PG, PH, PK, PL, PM, PN, PR, PS, PT, PW, PY, QA, RE, RO, RS, RU, RW, SA, SB, SC, SD, SE, SG, SH, SI, SJ, SK, SL, SM, SN, SO, SR, SS, ST, SV, SX, SY, SZ, TC, TD, TF, TG, TH, TJ, TK, TL, TM, TN, TO, TR, TT, TV, TW, TZ, UA, UG, UM, US, UY, UZ, VA, VC, VE, VG, VI, VN, VU, WF, WS, YE, YT, ZA, ZM, ZW
      --currency string                                                             Currency of the customer. Format must be ISO 4217; one of: , AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --customer-type string                                                        The type of the customer. It can have one of the following values:
                                                                                    - 'company': the customer is a company.
                                                                                    - 'individual': the customer is an individual.; one of: company, individual,
      --email string                                                                The email of the customer
      --external-id string                                                          The customer external unique identifier (provided by your own application)
      --external-salesforce-id string                                               The customer ID within the Salesforce system
      --finalize-zero-amount-invoice string                                         Specifies how invoices with a zero total amount should be handled:
                                                                                    - 'inherit': (Default) Follows the organization-level configuration.
                                                                                    - 'finalize': Invoices are issued and finalized even if the total amount is zero.
                                                                                    - 'skip': Invoices with a total amount of zero are not finalized.; one of: inherit, skip, finalize
      --firstname string                                                            First name of the customer
  -h, --help                                                                        help for create
      --idempotency-key string                                                      Idempotency key for safe mutation retries
      --input string                                                                Complete JSON request body or @file.json
      --integration-customers string                                                API field (array)
      --invoice-custom-section-codes string                                         List of unique codes identifying the invoice custom sections.
      --lastname string                                                             Last name of the customer
      --legal-name string                                                           The legal company name of the customer
      --legal-number string                                                         The legal company number of the customer
      --logo-url string                                                             The logo URL of the customer
      --metadata string                                                             Set of key-value pairs that you can attach to a customer. This can be useful for storing additional information about the customer in a structured format
      --name string                                                                 The full name of the customer (255 characters max).
      --net-payment-term string                                                     The net payment term, expressed in days, specifies the duration within which a customer is expected to remit payment after the invoice is finalized.
      --phone string                                                                The phone number of the customer
      --shipping-address-address-line1 string                                       The first line of the billing address
      --shipping-address-address-line2 string                                       The second line of the billing address
      --shipping-address-city string                                                The city of the customer's billing address
      --shipping-address-country string                                             Country code of the customer's billing address. Format must be ISO 3166 (alpha-2); one of: , AD, AE, AF, AG, AI, AL, AM, AO, AQ, AR, AS, AT, AU, AW, AX, AZ, BA, BB, BD, BE, BF, BG, BH, BI, BJ, BL, BM, BN, BO, BQ, BR, BS, BT, BV, BW, BY, BZ, CA, CC, CD, CF, CG, CH, CI, CK, CL, CM, CN, CO, CR, CU, CV, CW, CX, CY, CZ, DE, DJ, DK, DM, DO, DZ, EC, EE, EG, EH, ER, ES, ET, FI, FJ, FK, FM, FO, FR, GA, GB, GD, GE, GF, GG, GH, GI, GL, GM, GN, GP, GQ, GR, GS, GT, GU, GW, GY, HK, HM, HN, HR, HT, HU, ID, IE, IL, IM, IN, IO, IQ, IR, IS, IT, JE, JM, JO, JP, KE, KG, KH, KI, KM, KN, KP, KR, KW, KY, KZ, LA, LB, LC, LI, LK, LR, LS, LT, LU, LV, LY, MA, MC, MD, ME, MF, MG, MH, MK, ML, MM, MN, MO, MP, MQ, MR, MS, MT, MU, MV, MW, MX, MY, MZ, NA, NC, NE, NF, NG, NI, NL, NO, NP, NR, NU, NZ, OM, PA, PE, PF, PG, PH, PK, PL, PM, PN, PR, PS, PT, PW, PY, QA, RE, RO, RS, RU, RW, SA, SB, SC, SD, SE, SG, SH, SI, SJ, SK, SL, SM, SN, SO, SR, SS, ST, SV, SX, SY, SZ, TC, TD, TF, TG, TH, TJ, TK, TL, TM, TN, TO, TR, TT, TV, TW, TZ, UA, UG, UM, US, UY, UZ, VA, VC, VE, VG, VI, VN, VU, WF, WS, YE, YT, ZA, ZM, ZW
      --shipping-address-state string                                               The state of the customer's billing address
      --shipping-address-zipcode string                                             The zipcode of the customer's billing address
      --skip-invoice-custom-sections string                                         Set to 'true' to exclude all invoice custom sections from PDF generation for this customer only. False by default
      --state string                                                                The state of the customer's billing address
      --tax-codes string                                                            List of unique code used to identify the taxes.
      --tax-identification-number string                                            The tax identification number of the customer
      --timezone string                                                             The customer's timezone, used for billing purposes in their local time. Overrides the organization's timezone; one of: , UTC, Africa/Algiers, Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Johannesburg, Africa/Monrovia, Africa/Nairobi, America/Argentina/Buenos_Aires, America/Bogota, America/Caracas, America/Chicago, America/Chihuahua, America/Denver, America/Godthab, America/Guatemala, America/Guyana, America/Halifax, America/Indiana/Indianapolis, America/Juneau, America/La_Paz, America/Lima, America/Los_Angeles, America/Mazatlan, America/Mexico_City, America/Monterrey, America/Montevideo, America/New_York, America/Phoenix, America/Puerto_Rico, America/Regina, America/Santiago, America/Sao_Paulo, America/St_Johns, America/Tijuana, Asia/Almaty, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Chongqing, Asia/Colombo, Asia/Dhaka, Asia/Hong_Kong, Asia/Irkutsk, Asia/Jakarta, Asia/Jerusalem, Asia/Kabul, Asia/Kamchatka, Asia/Karachi, Asia/Kathmandu, Asia/Kolkata, Asia/Krasnoyarsk, Asia/Kuala_Lumpur, Asia/Kuwait, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Rangoon, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Srednekolymsk, Asia/Taipei, Asia/Tashkent, Asia/Tbilisi, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Urumqi, Asia/Vladivostok, Asia/Yakutsk, Asia/Yekaterinburg, Asia/Yerevan, Atlantic/Azores, Atlantic/Cape_Verde, Atlantic/South_Georgia, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Melbourne, Australia/Perth, Australia/Sydney, Europe/Amsterdam, Europe/Athens, Europe/Belgrade, Europe/Berlin, Europe/Bratislava, Europe/Brussels, Europe/Bucharest, Europe/Budapest, Europe/Copenhagen, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Kiev, Europe/Lisbon, Europe/Ljubljana, Europe/London, Europe/Madrid, Europe/Minsk, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Riga, Europe/Rome, Europe/Samara, Europe/Sarajevo, Europe/Skopje, Europe/Sofia, Europe/Stockholm, Europe/Tallinn, Europe/Vienna, Europe/Vilnius, Europe/Volgograd, Europe/Warsaw, Europe/Zagreb, Europe/Zurich, GMT+12, Pacific/Apia, Pacific/Auckland, Pacific/Chatham, Pacific/Fakaofo, Pacific/Fiji, Pacific/Guadalcanal, Pacific/Guam, Pacific/Honolulu, Pacific/Majuro, Pacific/Midway, Pacific/Noumea, Pacific/Pago_Pago, Pacific/Port_Moresby, Pacific/Tongatapu
      --url string                                                                  The custom website URL of the customer
      --zipcode string                                                              The zipcode of the customer's billing address
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

* [lago customers](lago_customers)	 - Manage Lago customers
