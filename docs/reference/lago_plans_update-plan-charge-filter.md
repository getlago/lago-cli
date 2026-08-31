## lago plans update-plan-charge-filter

Update a charge filter

### Synopsis

This endpoint updates a specific filter for a charge.

```
lago plans update-plan-charge-filter <code> <charge_code> <filter_id> [flags]
```

### Examples

```
  lago plans update-plan-charge-filter <code> <charge_code> <filter_id> --input @payload.json
```

### Options

```
      --cascade-updates string                               This field determines whether the changes made to the filter should be cascaded to the children plans. When set to 'true', the changes will be cascaded into children. Conversely, when set to 'false', the changes will only be applied to the plan itself. If not defined in the request, default value is 'false'.
  -h, --help                                                 help for update-plan-charge-filter
      --idempotency-key string                               Idempotency key for safe mutation retries
      --input string                                         Complete JSON request body or @file.json
      --invoice-display-name string                          Specifies the name that will be displayed on an invoice. If no value is set for this field, the values of the filter will be used as the default display name.
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
      --values string                                        List of possible filter values. The key and values must match one of the billable metric filters.
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
