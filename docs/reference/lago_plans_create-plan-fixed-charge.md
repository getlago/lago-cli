## lago plans create-plan-fixed-charge

Create a fixed charge

### Synopsis

This endpoint creates a new fixed charge for a specific plan.

```
lago plans create-plan-fixed-charge <code> [flags]
```

### Examples

```
  lago plans create-plan-fixed-charge <code> --input @payload.json
  lago plans create-plan-fixed-charge <code> --input @payload.json --output json  # full resource
```

### Options

```
      --add-on-code string                                   Unique code identifying an add-on. Either add_on_id or add_on_code is required.
      --add-on-id string                                     Unique identifier of the add-on. Either add_on_id or add_on_code is required.
      --apply-units-immediately string                       When set to 'true', the fixed charge units are applied immediately for active subscriptions. When set to 'false', the units are applied at the next billing period.
      --cascade-updates string                               This field determines whether the creation of the fixed charge should be cascaded to the children plans. When set to 'true', the fixed charge will be created in children plans. Conversely, when set to 'false', the fixed charge will only be created in the plan itself. If not defined in the request, default value is 'false'.
      --charge-model string                                  Specifies the pricing model used for the calculation of the fixed charge fee. It can be any of the following values:
                                                               - 'standard'
                                                               - 'graduated'
                                                               - 'volume'; one of: standard, graduated, volume
      --code string                                          Unique code identifying the fixed charge within the plan.
  -h, --help                                                 help for create-plan-fixed-charge
      --input string                                         Complete JSON request body or @file.json
      --invoice-display-name string                          Specifies the name that will be displayed on an invoice. If no value is set for this field, the name of the add-on will be used as the default display name.
      --pay-in-advance string                                This field determines the billing timing for this fixed charge. When set to 'true', the charge is due and invoiced immediately at the beginning of the billing period. When set to 'false', the charge is due and invoiced at the end of the billing period.
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
      --prorated string                                      Specifies whether a fixed charge is prorated based on the remaining number of days in the billing period or billed fully.

                                                             - If set to 'true', the charge is prorated based on the remaining days in the current billing period.
                                                             - If set to 'false', the charge is billed in full.
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

* [lago plans](lago_plans)	 - Manage Lago plans
