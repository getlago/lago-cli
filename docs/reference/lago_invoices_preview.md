## lago invoices preview

Create an invoice preview

### Synopsis

This endpoint is used for generating invoice preview.

```
lago invoices preview [flags]
```

### Examples

```
  lago invoices preview --input @payload.json
```

### Options

```
      --billing-entity-code string                       The code of the billing entity to which will be associated a customer if the external_id is not provided. If billing_entity_code is not provided, default billing_entity of organization will be used.
      --billing-time string                              The billing time for the subscription, which can be set as either 'anniversary' or 'calendar'. If not explicitly provided, it will default to 'calendar'. The billing time determines the timing of recurring billing cycles for the subscription. By specifying 'anniversary', the billing cycle will be based on the specific date the subscription started (billed fully), while 'calendar' sets the billing cycle at the first day of the week/month/year (billed with proration).; one of: calendar, anniversary
      --coupons string                                   API field (array)
      --customer-address-line1 string                    The first line of the billing address
      --customer-address-line2 string                    The second line of the billing address
      --customer-city string                             The city of the customer's billing address
      --customer-country string                          Country code of the customer's billing address. Format must be ISO 3166 (alpha-2); one of: AD, AE, AF, AG, AI, AL, AM, AO, AQ, AR, AS, AT, AU, AW, AX, AZ, BA, BB, BD, BE, BF, BG, BH, BI, BJ, BL, BM, BN, BO, BQ, BR, BS, BT, BV, BW, BY, BZ, CA, CC, CD, CF, CG, CH, CI, CK, CL, CM, CN, CO, CR, CU, CV, CW, CX, CY, CZ, DE, DJ, DK, DM, DO, DZ, EC, EE, EG, EH, ER, ES, ET, FI, FJ, FK, FM, FO, FR, GA, GB, GD, GE, GF, GG, GH, GI, GL, GM, GN, GP, GQ, GR, GS, GT, GU, GW, GY, HK, HM, HN, HR, HT, HU, ID, IE, IL, IM, IN, IO, IQ, IR, IS, IT, JE, JM, JO, JP, KE, KG, KH, KI, KM, KN, KP, KR, KW, KY, KZ, LA, LB, LC, LI, LK, LR, LS, LT, LU, LV, LY, MA, MC, MD, ME, MF, MG, MH, MK, ML, MM, MN, MO, MP, MQ, MR, MS, MT, MU, MV, MW, MX, MY, MZ, NA, NC, NE, NF, NG, NI, NL, NO, NP, NR, NU, NZ, OM, PA, PE, PF, PG, PH, PK, PL, PM, PN, PR, PS, PT, PW, PY, QA, RE, RO, RS, RU, RW, SA, SB, SC, SD, SE, SG, SH, SI, SJ, SK, SL, SM, SN, SO, SR, SS, ST, SV, SX, SY, SZ, TC, TD, TF, TG, TH, TJ, TK, TL, TM, TN, TO, TR, TT, TV, TW, TZ, UA, UG, UM, US, UY, UZ, VA, VC, VE, VG, VI, VN, VU, WF, WS, YE, YT, ZA, ZM, ZW
      --customer-currency string                         The currency of the invoice.; one of: AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GHS, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --customer-external-id string                      Unique identifier assigned to the customer in your application.
      --customer-integration-customers string            API field (array)
      --customer-name string                             The full name of the customer
      --customer-shipping-address-address-line1 string   The first line of the billing address
      --customer-shipping-address-address-line2 string   The second line of the billing address
      --customer-shipping-address-city string            The city of the customer's billing address
      --customer-shipping-address-country string         Country code of the customer's billing address. Format must be ISO 3166 (alpha-2); one of: , AD, AE, AF, AG, AI, AL, AM, AO, AQ, AR, AS, AT, AU, AW, AX, AZ, BA, BB, BD, BE, BF, BG, BH, BI, BJ, BL, BM, BN, BO, BQ, BR, BS, BT, BV, BW, BY, BZ, CA, CC, CD, CF, CG, CH, CI, CK, CL, CM, CN, CO, CR, CU, CV, CW, CX, CY, CZ, DE, DJ, DK, DM, DO, DZ, EC, EE, EG, EH, ER, ES, ET, FI, FJ, FK, FM, FO, FR, GA, GB, GD, GE, GF, GG, GH, GI, GL, GM, GN, GP, GQ, GR, GS, GT, GU, GW, GY, HK, HM, HN, HR, HT, HU, ID, IE, IL, IM, IN, IO, IQ, IR, IS, IT, JE, JM, JO, JP, KE, KG, KH, KI, KM, KN, KP, KR, KW, KY, KZ, LA, LB, LC, LI, LK, LR, LS, LT, LU, LV, LY, MA, MC, MD, ME, MF, MG, MH, MK, ML, MM, MN, MO, MP, MQ, MR, MS, MT, MU, MV, MW, MX, MY, MZ, NA, NC, NE, NF, NG, NI, NL, NO, NP, NR, NU, NZ, OM, PA, PE, PF, PG, PH, PK, PL, PM, PN, PR, PS, PT, PW, PY, QA, RE, RO, RS, RU, RW, SA, SB, SC, SD, SE, SG, SH, SI, SJ, SK, SL, SM, SN, SO, SR, SS, ST, SV, SX, SY, SZ, TC, TD, TF, TG, TH, TJ, TK, TL, TM, TN, TO, TR, TT, TV, TW, TZ, UA, UG, UM, US, UY, UZ, VA, VC, VE, VG, VI, VN, VU, WF, WS, YE, YT, ZA, ZM, ZW
      --customer-shipping-address-state string           The state of the customer's billing address
      --customer-shipping-address-zipcode string         The zipcode of the customer's billing address
      --customer-state string                            The state of the customer's billing address
      --customer-tax-identification-number string        The tax identification number of the customer
      --customer-timezone string                         The customer's timezone, used for billing purposes in their local time. Overrides the organization's timezone; one of: UTC, Africa/Algiers, Africa/Cairo, Africa/Casablanca, Africa/Harare, Africa/Johannesburg, Africa/Monrovia, Africa/Nairobi, America/Argentina/Buenos_Aires, America/Bogota, America/Caracas, America/Chicago, America/Chihuahua, America/Denver, America/Guatemala, America/Guyana, America/Halifax, America/Indiana/Indianapolis, America/Juneau, America/La_Paz, America/Lima, America/Los_Angeles, America/Mazatlan, America/Mexico_City, America/Monterrey, America/Montevideo, America/New_York, America/Nuuk, America/Phoenix, America/Puerto_Rico, America/Regina, America/Santiago, America/Sao_Paulo, America/St_Johns, America/Tijuana, Asia/Almaty, Asia/Baghdad, Asia/Baku, Asia/Bangkok, Asia/Chongqing, Asia/Colombo, Asia/Dhaka, Asia/Hong_Kong, Asia/Irkutsk, Asia/Jakarta, Asia/Jerusalem, Asia/Kabul, Asia/Kamchatka, Asia/Karachi, Asia/Kathmandu, Asia/Kolkata, Asia/Krasnoyarsk, Asia/Kuala_Lumpur, Asia/Kuwait, Asia/Magadan, Asia/Muscat, Asia/Novosibirsk, Asia/Riyadh, Asia/Seoul, Asia/Shanghai, Asia/Singapore, Asia/Srednekolymsk, Asia/Taipei, Asia/Tashkent, Asia/Tbilisi, Asia/Tehran, Asia/Tokyo, Asia/Ulaanbaatar, Asia/Urumqi, Asia/Vladivostok, Asia/Yakutsk, Asia/Yangon, Asia/Yekaterinburg, Asia/Yerevan, Atlantic/Azores, Atlantic/Cape_Verde, Atlantic/South_Georgia, Australia/Adelaide, Australia/Brisbane, Australia/Darwin, Australia/Hobart, Australia/Melbourne, Australia/Perth, Australia/Sydney, Europe/Amsterdam, Europe/Athens, Europe/Belgrade, Europe/Berlin, Europe/Bratislava, Europe/Brussels, Europe/Bucharest, Europe/Budapest, Europe/Copenhagen, Europe/Dublin, Europe/Helsinki, Europe/Istanbul, Europe/Kaliningrad, Europe/Kyiv, Europe/Lisbon, Europe/Ljubljana, Europe/London, Europe/Madrid, Europe/Minsk, Europe/Moscow, Europe/Paris, Europe/Prague, Europe/Riga, Europe/Rome, Europe/Samara, Europe/Sarajevo, Europe/Skopje, Europe/Sofia, Europe/Stockholm, Europe/Tallinn, Europe/Vienna, Europe/Vilnius, Europe/Volgograd, Europe/Warsaw, Europe/Zagreb, Europe/Zurich, GMT+12, Pacific/Apia, Pacific/Auckland, Pacific/Chatham, Pacific/Fakaofo, Pacific/Fiji, Pacific/Guadalcanal, Pacific/Guam, Pacific/Honolulu, Pacific/Majuro, Pacific/Midway, Pacific/Noumea, Pacific/Pago_Pago, Pacific/Port_Moresby, Pacific/Tongatapu
  -h, --help                                             help for preview
      --input string                                     Complete JSON request body or @file.json
      --plan-code string                                 The code of the plan. It serves as a unique identifier associated with a particular plan. The code is typically used for internal or system-level identification purposes, like assigning a subscription, for instance.
      --subscription-at string                           The anniversary date and time of the initial subscription. This date serves as the basis for billing subscriptions with 'anniversary' billing time. The 'anniversary_date' should be provided in ISO 8601 datetime format and expressed in Coordinated Universal Time (UTC).
      --subscriptions-external-ids string                API field (array)
      --subscriptions-plan-code string                   The code of the plan. It serves as a unique identifier associated with the new plan to which the customer is upgrading or downgrading.
      --subscriptions-terminated-at string               The date and time when the subscription is terminated. It should defined in the future.
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
