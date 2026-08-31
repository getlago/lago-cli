## lago billing-entities update

Update a billing entity

### Synopsis

This endpoint is used to update an existing billing entity

```
lago billing-entities update <code> [flags]
```

### Examples

```
  lago billing-entities update <code> --input @payload.json
```

### Options

```
      --address-line1 string                                                        The first line of the billing address
      --address-line2 string                                                        The second line of the billing address
      --billing-configuration-document-locale string                                The language of the documents generated for this billing entity
      --billing-configuration-invoice-footer string                                 The footer text to be displayed on invoices for this billing entity
      --billing-configuration-invoice-grace-period string                           The grace period (in days) for invoice finalization
      --billing-configuration-subscription-invoice-issuing-date-adjustment string   The logic applied on top of the subscription_invoice_issuing_date_anchor rule. You can opt to use the invoice finalization date, that includes any configured grace period.; one of: align_with_finalization_date, keep_anchor
      --billing-configuration-subscription-invoice-issuing-date-anchor string       Defines whether the issuing date follows the current billing period's end date or the next period starting date.; one of: current_period_end, next_period_start
      --city string                                                                 The city of the billing address
      --country string                                                              The country code of the billing address; one of: , AD, AE, AF, AG, AI, AL, AM, AO, AQ, AR, AS, AT, AU, AW, AX, AZ, BA, BB, BD, BE, BF, BG, BH, BI, BJ, BL, BM, BN, BO, BQ, BR, BS, BT, BV, BW, BY, BZ, CA, CC, CD, CF, CG, CH, CI, CK, CL, CM, CN, CO, CR, CU, CV, CW, CX, CY, CZ, DE, DJ, DK, DM, DO, DZ, EC, EE, EG, EH, ER, ES, ET, FI, FJ, FK, FM, FO, FR, GA, GB, GD, GE, GF, GG, GH, GI, GL, GM, GN, GP, GQ, GR, GS, GT, GU, GW, GY, HK, HM, HN, HR, HT, HU, ID, IE, IL, IM, IN, IO, IQ, IR, IS, IT, JE, JM, JO, JP, KE, KG, KH, KI, KM, KN, KP, KR, KW, KY, KZ, LA, LB, LC, LI, LK, LR, LS, LT, LU, LV, LY, MA, MC, MD, ME, MF, MG, MH, MK, ML, MM, MN, MO, MP, MQ, MR, MS, MT, MU, MV, MW, MX, MY, MZ, NA, NC, NE, NF, NG, NI, NL, NO, NP, NR, NU, NZ, OM, PA, PE, PF, PG, PH, PK, PL, PM, PN, PR, PS, PT, PW, PY, QA, RE, RO, RS, RU, RW, SA, SB, SC, SD, SE, SG, SH, SI, SJ, SK, SL, SM, SN, SO, SR, SS, ST, SV, SX, SY, SZ, TC, TD, TF, TG, TH, TJ, TK, TL, TM, TN, TO, TR, TT, TV, TW, TZ, UA, UG, UM, US, UY, UZ, VA, VC, VE, VG, VI, VN, VU, WF, WS, YE, YT, ZA, ZM, ZW
      --default-currency string                                                     The default currency of the billing entity; one of: AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GHS, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --document-number-prefix string                                               The prefix used in document numbers for this billing entity
      --document-numbering string                                                   The type of document numbering for this billing entity:
                                                                                    - 'per_customer': document numbers are unique per customer
                                                                                    - 'per_billing_entity': document numbers are unique per billing entity; one of: per_customer, per_billing_entity
      --einvoicing string                                                           Whether e-invoicing is enabled for this billing entity
      --email string                                                                The email address of the billing entity
      --email-settings string                                                       The email notification settings for this billing entity
      --eu-tax-management string                                                    Whether EU tax management is enabled for this billing entity
      --finalize-zero-amount-invoice string                                         Whether to finalize invoices with zero amount for this billing entity
  -h, --help                                                                        help for update
      --input string                                                                Complete JSON request body or @file.json
      --invoice-custom-section-codes string                                         The codes of the invoice custom section that should be associated with this billing entity
      --legal-name string                                                           The legal name of the billing entity
      --legal-number string                                                         The legal registration number of the billing entity
      --logo string                                                                 The base64 encoded logo image for the billing entity. Sending "null" will remove the logo, if any exist.
      --name string                                                                 The name of the billing entity
      --net-payment-term string                                                     The net payment term (in days) for this billing entity
      --phone string                                                                The phone number of the billing entity
      --state string                                                                The state of the billing address
      --tax-codes string                                                            The tax codes that should be associated with this billing entity
      --tax-identification-number string                                            The tax identification number of the billing entity
      --timezone string                                                             The timezone of the billing entity; one of: UTC, Africa/Algiers, Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Johannesburg, Africa/Monrovia, Africa/Nairobi, America/Argentina/Buenos_Aires, America/Bogota, America/Caracas, America/Chicago, America/Chihuahua, America/Denver, America/Guatemala, America/Guyana, America/Halifax, America/Indiana/Indianapolis, America/Juneau, America/La_Paz, America/Lima, America/Los_Angeles, America/Mazatlan, America/Mexico_City, America/Monterrey, America/Montevideo, America/New_York, America/Nuuk, America/Phoenix, America/Puerto_Rico, America/Regina, America/Santiago, America/Sao_Paulo, America/St_Johns, America/Tijuana, Asia/Almaty, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Chongqing, Asia/Colombo, Asia/Dhaka, Asia/Hong_Kong, Asia/Irkutsk, Asia/Jakarta, Asia/Jerusalem, Asia/Kabul, Asia/Kamchatka, Asia/Karachi, Asia/Kathmandu, Asia/Kolkata, Asia/Krasnoyarsk, Asia/Kuala_Lumpur, Asia/Kuwait, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Srednekolymsk, Asia/Taipei, Asia/Tashkent, Asia/Tbilisi, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Urumqi, Asia/Vladivostok, Asia/Yakutsk, Asia/Yangon, Asia/Yekaterinburg, Asia/Yerevan, Atlantic/Azores, Atlantic/Cape_Verde, Atlantic/South_Georgia, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Melbourne, Australia/Perth, Australia/Sydney, Europe/Amsterdam, Europe/Athens, Europe/Belgrade, Europe/Berlin, Europe/Bratislava, Europe/Brussels, Europe/Bucharest, Europe/Budapest, Europe/Copenhagen, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Kyiv, Europe/Lisbon, Europe/Ljubljana, Europe/London, Europe/Madrid, Europe/Minsk, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Riga, Europe/Rome, Europe/Samara, Europe/Sarajevo, Europe/Skopje, Europe/Sofia, Europe/Stockholm, Europe/Tallinn, Europe/Vienna, Europe/Vilnius, Europe/Volgograd, Europe/Warsaw, Europe/Zagreb, Europe/Zurich, GMT+12, Pacific/Apia, Pacific/Auckland, Pacific/Chatham, Pacific/Fakaofo, Pacific/Fiji, Pacific/Guadalcanal, Pacific/Guam, Pacific/Honolulu, Pacific/Majuro, Pacific/Midway, Pacific/Noumea, Pacific/Pago_Pago, Pacific/Port_Moresby, Pacific/Tongatapu
      --zipcode string                                                              The zipcode of the billing address
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

* [lago billing-entities](lago_billing-entities)	 - Manage Lago billing entities
