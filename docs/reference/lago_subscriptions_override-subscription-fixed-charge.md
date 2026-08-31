## lago subscriptions override-subscription-fixed-charge

Override a fixed charge for a subscription

### Synopsis

This endpoint creates or updates a fixed charge override for a subscription, without affecting the original plan fixed charge.
When the body contains only `units` (and optionally `apply_units_immediately`), the change is recorded as a per-subscription unit override. When the body also sets `invoice_display_name`, `properties`, or `tax_codes`, it is applied as a full plan override for the subscription instead.
With `apply_units_immediately: true` on a pay-in-advance fixed charge the change is billed mid-period; otherwise it takes effect at the next billing period.
This is a premium feature.

```
lago subscriptions override-subscription-fixed-charge <external_id> <fixed_charge_code> [flags]
```

### Examples

```
  lago subscriptions override-subscription-fixed-charge <external_id> <fixed_charge_code> --input @payload.json
```

### Options

```
      --apply-units-immediately string                       When set to 'true', the fixed charge units are applied immediately for active subscriptions. When set to 'false', the units are applied at the next billing period.
  -h, --help                                                 help for override-subscription-fixed-charge
      --input string                                         Complete JSON request body or @file.json
      --invoice-display-name string                          Specifies the name that will be displayed on an invoice. If no value is set for this field, the name of the add-on will be used as the default display name.
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
      --properties-pricing-group-keys string                 The list of event properties that are used to group the events on the invoice.
      --properties-rate string                               The percentage rate that is applied to the amount of each transaction for a 'percentage' charge model. It is expressed as a decimal value.
      --properties-volume-ranges string                      Volume ranges, sorted from bottom to top tiers, used for a 'volume' charge model.
      --subscription-status string                           Filter by subscription status. When provided, the subscription is looked up with this status instead of the default 'active' status. Possible values are 'pending', 'active', 'terminated', or 'canceled'.; one of: pending, active, terminated, canceled
      --tax-codes string                                     List of unique code used to identify the taxes.
      --units string                                         The quantity of units for the fixed charge.
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
