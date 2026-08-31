## lago subscriptions create

Assign a plan to a customer

### Synopsis

This endpoint assigns a plan to a customer, creating or modifying a subscription. Ideal for initial subscriptions or plan changes (upgrades/downgrades).

```
lago subscriptions create [flags]
```

### Examples

```
  lago subscriptions create --input @payload.json
```

### Options

```
      --authorization-amount-cents string                                            The amount of the authorization in cents.
      --authorization-amount-currency string                                         The currency of the authorization.
  -h, --help                                                                         help for create
      --idempotency-key string                                                       Idempotency key for safe mutation retries
      --input string                                                                 Complete JSON request body or @file.json
      --subscription-activation-rules string                                         Optional list of activation rules that gate the subscription activation. When a 'payment' rule is provided and the plan is paid in advance (and the subscription is not in a trial), the subscription is created in the 'incomplete' state and is only activated once the gating payment succeeds. If the payment fails or the rule's 'timeout_hours' elapses, the subscription is canceled.
      --subscription-billing-entity-code string                                      The code of the billing entity to be used for the subscription. If not provided, the default billing entity will be used.
      --subscription-billing-time string                                             The billing time for the subscription, which can be set as either 'anniversary' or 'calendar'. If not explicitly provided, it will default to 'calendar'. The billing time determines the timing of recurring billing cycles for the subscription. By specifying 'anniversary', the billing cycle will be based on the specific date the subscription started (billed fully), while 'calendar' sets the billing cycle at the first day of the week/month/year (billed with proration).; one of: calendar, anniversary
      --subscription-consolidate-invoice string                                      Defines whether this subscription should be grouped with other subscriptions of the same customer when generating recurring invoices.

                                                                                     - 'true' (default): the subscription is included in the customer's standard invoice grouping (by billing entity, currency and payment method).
                                                                                     - 'false': the subscription is excluded from consolidation and always billed on its own dedicated invoice.
      --subscription-ending-at string                                                The effective end date of the subscription. If this field is set to null, the subscription will automatically renew. This date should be provided in ISO 8601 datetime format, and use Coordinated Universal Time (UTC).
      --subscription-external-customer-id string                                     The customer external unique identifier (provided by your own application)
      --subscription-external-id string                                              The unique external identifier for the subscription. This identifier serves as an idempotency key, ensuring that each subscription is unique.
      --subscription-invoice-custom-section-invoice-custom-section-codes string      List of unique codes identifying the invoice custom sections to apply. These override the default invoice custom sections configured at the customer or billing entity level.
      --subscription-invoice-custom-section-skip-invoice-custom-sections string      Set to 'true' to exclude all invoice custom sections from PDF generation for invoices related to this resource. When 'true', 'invoice_custom_section_codes' is ignored.
      --subscription-name string                                                     The display name of the subscription on an invoice. This field allows for customization of the subscription's name for billing purposes, especially useful when a single customer has multiple subscriptions using the same plan.
      --subscription-payment-method-payment-method-id string                         The unique identifier of the payment method (required when using a specific provider payment method).
      --subscription-payment-method-payment-method-type string                       The type of payment method to use.; one of: provider, manual
      --subscription-plan-code string                                                The unique code representing the plan to be attached to the customer. This code must correspond to the 'code' property of one of the active plans.
      --subscription-plan-overrides-amount-cents string                              The base cost of the plan, excluding any applicable taxes, that is billed on a recurring basis. This value is defined at 0 if your plan is a pay-as-you-go plan.
      --subscription-plan-overrides-amount-currency string                           The currency of the plan. It indicates the monetary unit in which the plan's cost, including taxes and usage-based charges, is expressed.; one of: AED, AFN, ALL, AMD, ANG, AOA, ARS, AUD, AWG, AZN, BAM, BBD, BDT, BGN, BIF, BMD, BND, BOB, BRL, BSD, BWP, BYN, BZD, CAD, CDF, CHF, CLF, CLP, CNY, COP, CRC, CVE, CZK, DJF, DKK, DOP, DZD, EGP, ETB, EUR, FJD, FKP, GBP, GEL, GHS, GIP, GMD, GNF, GTQ, GYD, HKD, HNL, HRK, HTG, HUF, IDR, ILS, INR, ISK, JMD, JPY, KES, KGS, KHR, KMF, KRW, KYD, KZT, LAK, LBP, LKR, LRD, LSL, MAD, MDL, MGA, MKD, MMK, MNT, MOP, MRO, MUR, MVR, MWK, MXN, MYR, MZN, NAD, NGN, NIO, NOK, NPR, NZD, PAB, PEN, PGK, PHP, PKR, PLN, PYG, QAR, RON, RSD, RUB, RWF, SAR, SBD, SCR, SEK, SGD, SHP, SLL, SOS, SRD, STD, SZL, THB, TJS, TOP, TRY, TTD, TWD, TZS, UAH, UGX, USD, UYU, UZS, VND, VUV, WST, XAF, XCD, XOF, XPF, YER, ZAR, ZMW
      --subscription-plan-overrides-charges string                                   Additional usage-based charges for this plan.
      --subscription-plan-overrides-description string                               The description on the plan.
      --subscription-plan-overrides-fixed-charges string                             Fixed charge overrides for the subscription. When 'plan_overrides' contains only 'fixed_charges' and every entry contains only 'id', 'units', and optionally 'apply_units_immediately', the units are recorded as a per-subscription override without creating a plan override, and subscription-scoped reads return these units. If any entry carries other fields, or 'plan_overrides' contains any other key, the request is applied as a full plan override for the subscription instead.
      --subscription-plan-overrides-invoice-display-name string                      Specifies the name that will be displayed on an invoice. If no value is set for this field, the name of the plan will be used as the default display name.
      --subscription-plan-overrides-metadata string                                  Metadata to set as key-value pairs. Keys are strings (max 100 characters), values can be strings (max 255 characters) or null.
      --subscription-plan-overrides-minimum-commitment-amount-cents string           The amount of the minimum commitment in cents.
      --subscription-plan-overrides-minimum-commitment-invoice-display-name string   Specifies the name that will be displayed on an invoice. If no value is set for this field, the default name will be used as the display name.
      --subscription-plan-overrides-minimum-commitment-tax-codes string              List of unique code used to identify the taxes.
      --subscription-plan-overrides-name string                                      The name of the plan.
      --subscription-plan-overrides-tax-codes string                                 List of unique code used to identify the taxes.
      --subscription-plan-overrides-trial-period string                              The duration in days during which the base cost of the plan is offered for free.
      --subscription-plan-overrides-usage-thresholds string                          **Deprecated.** Managing usage thresholds through 'plan_overrides' is deprecated and its behavior is inconsistent: an empty array does not reliably clear thresholds (it is ignored the first time a subscription is overridden), and it never affects thresholds inherited from a parent plan. Use the top-level 'usage_thresholds' array on the subscription instead, or 'progressive_billing_disabled' to switch progressive billing off.
      --subscription-progressive-billing-disabled string                             Disables progressive billing for this subscription, regardless of any usage thresholds defined on the plan (including thresholds inherited from a parent plan when the subscription runs on a plan override).

                                                                                     Set to 'true' to switch progressive billing off for this subscription specifically, without changing the plan's thresholds.
      --subscription-purchase-order-number string                                    The purchase order number associated with the subscription. It will be added to invoices generated for this subscription.
      --subscription-subscription-at string                                          The start date for the subscription, allowing for the creation of subscriptions that can begin in the past or future. Please note that it cannot be used to update the start date of a pending subscription or schedule an upgrade/downgrade. The start_date should be provided in ISO 8601 datetime format and expressed in Coordinated Universal Time (UTC).
      --subscription-usage-thresholds string                                         Usage thresholds managed directly on the subscription. This is the recommended way to set progressive billing thresholds per subscription; when present, they take precedence over the thresholds defined on the plan.

                                                                                     This field cannot be combined with 'plan_overrides.usage_thresholds' in the same request.
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

* [lago subscriptions](lago_subscriptions)	 - Manage Lago subscriptions
