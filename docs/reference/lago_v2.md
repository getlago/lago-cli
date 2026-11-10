## lago v2

Lago API v2 (beta): product catalog, contracts and v2 resources

### Synopsis

Commands generated from the Lago API v2 document (2.0.0-beta).

API v2 adds the product catalog (products, product categories, product filters, rate cards and rates), catalog plans, contracts and their applied rate cards and rate phases, and native v2 versions of billing resources. Lists use cursor pagination: pass --after with the previous response's meta.next_cursor, or --all to walk every page. Requests go to /api/v2 on the profile's API host.

```
lago v2 [flags]
```

### Examples

```
  lago v2 products list
  lago v2 products create --code compute --name Compute --product-type fixed
  lago v2 rate-cards list --all
```

### Options

```
  -h, --help   help for v2
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

* [lago](lago)	 - The official CLI for Lago billing
* [lago v2 activity-logs](lago_v2_activity-logs)	 - Manage Lago activity logs
* [lago v2 api-logs](lago_v2_api-logs)	 - Manage Lago api logs
* [lago v2 billable-metrics](lago_v2_billable-metrics)	 - Manage Lago billable metrics
* [lago v2 billing-entities](lago_v2_billing-entities)	 - Manage Lago billing entities
* [lago v2 catalog-plans](lago_v2_catalog-plans)	 - Manage Lago catalog plans
* [lago v2 contract-applied-rate-cards](lago_v2_contract-applied-rate-cards)	 - Manage Lago contract applied rate cards
* [lago v2 contract-rate-phases](lago_v2_contract-rate-phases)	 - Manage Lago contract rate phases
* [lago v2 contracts](lago_v2_contracts)	 - Manage Lago contracts
* [lago v2 credit-notes](lago_v2_credit-notes)	 - Manage Lago credit notes
* [lago v2 customers](lago_v2_customers)	 - Manage Lago customers
* [lago v2 events](lago_v2_events)	 - Manage Lago events
* [lago v2 fees](lago_v2_fees)	 - Manage Lago fees
* [lago v2 invoices](lago_v2_invoices)	 - Manage Lago invoices
* [lago v2 organizations](lago_v2_organizations)	 - Manage Lago organizations
* [lago v2 payment-receipts](lago_v2_payment-receipts)	 - Manage Lago payment receipts
* [lago v2 payment-requests](lago_v2_payment-requests)	 - Manage Lago payment requests
* [lago v2 payments](lago_v2_payments)	 - Manage Lago payments
* [lago v2 plan-applied-rate-cards](lago_v2_plan-applied-rate-cards)	 - Manage Lago plan applied rate cards
* [lago v2 plan-rate-phases](lago_v2_plan-rate-phases)	 - Manage Lago plan rate phases
* [lago v2 product-categories](lago_v2_product-categories)	 - Manage Lago product categories
* [lago v2 product-filters](lago_v2_product-filters)	 - Manage Lago product filters
* [lago v2 products](lago_v2_products)	 - Manage Lago products
* [lago v2 rate-card-rates](lago_v2_rate-card-rates)	 - Manage Lago rate card rates
* [lago v2 rate-cards](lago_v2_rate-cards)	 - Manage Lago rate cards
* [lago v2 security-logs](lago_v2_security-logs)	 - Manage Lago security logs
* [lago v2 taxes](lago_v2_taxes)	 - Manage Lago taxes
* [lago v2 webhook-endpoints](lago_v2_webhook-endpoints)	 - Manage Lago webhook endpoints
