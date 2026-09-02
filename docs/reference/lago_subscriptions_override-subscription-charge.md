## lago subscriptions override-subscription-charge

Override a charge for a subscription

### Synopsis

This endpoint creates or updates a charge override for a specific subscription.
If the subscription does not have a plan override yet, one will be created automatically.
The charge override allows customizing specific charge properties (invoice_display_name, min_amount_cents, properties, filters, taxes, applied_pricing_unit) without affecting the original plan charge.
This is a premium feature.

```
lago subscriptions override-subscription-charge <external_id> <charge_code> [flags]
```

### Examples

```
  lago subscriptions override-subscription-charge <external_id> <charge_code> --input @payload.json
```

### Options

```
      --applied-pricing-unit-code string                     The code of the pricing unit.
      --applied-pricing-unit-conversion-rate string          The conversion rate from pricing units to the plan's currency.
                                                             This rate determines how many currency units (in the plan's base currency) equal one pricing unit.
                                                             For example, if the plan uses USD and the conversion rate is 0.5, then 1 pricing unit = $0.50.
      --filters string                                       List of filters used to apply differentiated pricing based on additional event properties.
  -h, --help                                                 help for override-subscription-charge
      --input string                                         Complete JSON request body or @file.json
      --invoice-display-name string                          Specifies the name that will be displayed on an invoice. If no value is set for this field, the name of the actual charge will be used as the default display name.
      --min-amount-cents string                              The minimum spending amount required for the charge, measured in cents and excluding any applicable taxes. It indicates the minimum amount that needs to be charged for each billing period.
      --properties-amount string                             - The unit price, excluding tax, for a 'standard' charge model. It is expressed as a decimal value.
                                                             - The amount, excluding tax, for a complete set of units in a 'package' charge model. It is expressed as a decimal value.
      --properties-fixed-amount string                       The fixed fee that is applied to each transaction for a 'percentage' charge model. It is expressed as a decimal value.
      --properties-free-units string                         The quantity of units that are provided free of charge for each billing period in a 'package' charge model. This field specifies the number of units that customers can use without incurring any additional cost during each billing cycle.
      --properties-free-units-per-events string              The count of transactions that are not impacted by the 'percentage' rate and fixed fee in a percentage charge model. This field indicates the number of transactions that are exempt from the calculation of charges based on the specified percentage rate and fixed fee.
      --properties-free-units-per-total-aggregation string   The transaction amount that is not impacted by the 'percentage' rate and fixed fee in a percentage charge model. This field indicates the portion of the transaction amount that is exempt from the calculation of charges based on the specified percentage rate and fixed fee.
      --properties-graduated-percentage-ranges string        Graduated percentage ranges, sorted from bottom to top tiers, used for a 'graduated_percentage' charge model.
      --properties-graduated-ranges string                   Graduated ranges, sorted from bottom to top tiers, used for a 'graduated' charge model.
      --properties-grouped-by string                         **Deprecated.** Replaced by 'pricing_group_keys'.
                                                             The list of event properties that are used to group the events on the invoice for a 'standard' charge model.
      --properties-package-size string                       The quantity of units included in each pack or set for a 'package' charge model. It indicates the number of units that are bundled together as a single package or set within the pricing structure.
      --properties-per-transaction-max-amount string         Specifies the maximum allowable spending for a single transaction. Working as a transaction cap.
      --properties-per-transaction-min-amount string         Specifies the minimum allowable spending for a single transaction. Working as a transaction floor.
      --properties-presentation-group-keys string            Groups usage into sub-items on invoices for display only, without affecting pricing or aggregation.
      --properties-pricing-group-keys string                 The list of event properties that are used to group the events on the invoice.
      --properties-rate string                               The percentage rate that is applied to the amount of each transaction for a 'percentage' charge model. It is expressed as a decimal value.
      --properties-volume-ranges string                      Volume ranges, sorted from bottom to top tiers, used for a 'volume' charge model.
      --subscription-status string                           Filter by subscription status. When provided, the subscription is looked up with this status instead of the default 'active' status. Possible values are 'pending', 'active', 'terminated', or 'canceled'.; one of: pending, active, terminated, canceled
      --tax-codes string                                     List of unique code used to identify the taxes.
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
