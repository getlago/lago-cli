## lago plans create-plan-charge

Create a charge

### Synopsis

This endpoint creates a new charge for a specific plan.

```
lago plans create-plan-charge <code> [flags]
```

### Examples

```
  lago plans create-plan-charge <code> --input @payload.json
  lago plans create-plan-charge <code> --input @payload.json --output json  # full resource
```

### Options

```
      --accepts-target-wallet string                         Specifies whether the charge accepts a target wallet for prepaid credits deduction.
                                                             When set to 'true', events may include a 'target_wallet_code' in their 'properties' to direct credit deduction to a specific wallet.
                                                             If no 'target_wallet_code' is provided on the event, the standard wallet selection logic is applied.
                                                             If the 'target_wallet_code' does not match an existing wallet, no credits will be applied for that event.

                                                             This field requires a premium integration.
      --applied-pricing-unit-code string                     The code of the pricing unit.
      --applied-pricing-unit-conversion-rate string          The conversion rate from pricing units to the plan's currency.
                                                             This rate determines how many currency units (in the plan's base currency) equal one pricing unit.
                                                             For example, if the plan uses USD and the conversion rate is 0.5, then 1 pricing unit = $0.50.
      --billable-metric-id string                            Unique identifier of the billable metric created by Lago.
      --cascade-updates string                               This field determines whether the creation of the charge should be cascaded to the children plans. When set to 'true', the charge will be created in children plans. Conversely, when set to 'false', the charge will only be created in the plan itself. If not defined in the request, default value is 'false'.
      --charge-model string                                  Specifies the pricing model used for the calculation of the final fee. It can be any of the following values:
                                                               - ['dynamic'](https://docs.getlago.com/guide/plans/charges/charge-models/dynamic)
                                                               - ['graduated_percentage'](https://docs.getlago.com/guide/plans/charges/charge-models/graduated-percentage)
                                                               - ['graduated'](https://docs.getlago.com/guide/plans/charges/charge-models/graduated)
                                                               - ['package'](https://docs.getlago.com/guide/plans/charges/charge-models/package)
                                                               - ['percentage'](https://docs.getlago.com/guide/plans/charges/charge-models/percentage)
                                                               - ['standard'](https://docs.getlago.com/guide/plans/charges/charge-models/standard)
                                                               - ['volume'](https://docs.getlago.com/guide/plans/charges/charge-models/volume); one of: dynamic, graduated, graduated_percentage, package, percentage, standard, volume
      --code string                                          Unique code identifying the charge within the plan.
      --filters string                                       List of filters used to apply differentiated pricing based on additional event properties.
  -h, --help                                                 help for create-plan-charge
      --input string                                         Complete JSON request body or @file.json
      --invoice-display-name string                          Specifies the name that will be displayed on an invoice. If no value is set for this field, the name of the actual charge will be used as the default display name.
      --invoiceable string                                   This field specifies whether the charge should be included in a proper invoice. If set to false, no invoice will be issued for this charge. You can only set it to 'false' when 'pay_in_advance' is 'true'.
      --min-amount-cents string                              The minimum spending amount required for the charge, measured in cents and excluding any applicable taxes. It indicates the minimum amount that needs to be charged for each billing period.
      --pay-in-advance string                                This field determines the billing timing for this specific usage-based charge. When set to 'true', the charge is due and invoiced immediately. Conversely, when set to false, the charge is due and invoiced at the end of each billing period.
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
      --prorated string                                      Specifies whether a charge is prorated based on the remaining number of days in the billing period or billed fully.

                                                             - If set to 'true', the charge is prorated based on the remaining days in the current billing period.
                                                             - If set to 'false', the charge is billed in full.
                                                             - If not defined in the request, default value is 'false'.
      --regroup-paid-fees string                             This setting can only be configured if 'pay_in_advance' is 'true' and 'invoiceable' is 'false'.
                                                             This field determines whether and when the charge fee should be included in
                                                             the invoice. If 'null', no invoice will be issued for this charge fee.
                                                             If 'invoice', an invoice will be generated at the end of the period,
                                                             consolidating all charge fees with a succeeded payment status.; one of: , invoice
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

* [lago plans](lago_plans)	 - Manage Lago plans
