## lago customers attributed-usage

Retrieve customer attributed usage

### Synopsis

This endpoint retrieves the usage of an active subscription broken down by the values of a usage attribution type, one level of the account tree at a time. Only charges whose billable metric uses the `count_agg` or `sum_agg` aggregation and is not recurring are included. By default, the usage covers the current billing period, narrowed to its last 31 days when the period is longer.

```
lago customers attributed-usage <external_customer_id> [flags]
```

### Examples

```
  lago customers attributed-usage <external_customer_id>
```

### Options

```
      --all                               Fetch every page
      --basis string                      Basis of the breakdown. With 'units', no amount is computed and values are ranked by units, summed across the selected charges. With 'amount', amounts are computed and values are ranked by amount.; one of: units, amount
      --charge-codes string               Codes of the charges of the plan to include. All the supported charges are included when it is not sent.
      --external-subscription-id string   The unique identifier of the subscription within your application. The subscription must be active.
      --filters string                    Values of usage attribution types to filter the usage by, keyed by type code, e.g. 'filters[department]=rnd&filters[model][]=opus&filters[model][]=haiku'. Each type accepts up to 20 values, and up to 3 'flat' types can be filtered on. More than 20 values for a type returns a 'too_many_values' validation error.
      --from-datetime string              Start of the period, in ISO 8601 format. It must be sent together with 'to_datetime', and the period cannot be longer than 31 days. It is moved to the start of the subscription when earlier.
      --group-by string                   Code of the usage attribution type to group the usage by. When the type has a parent, 'filters' must include a value of the parent type, so the tree is read one level at a time. Lago tracks up to 100,000 values per type for each customer. Once the type reaches this limit, grouping by it returns a 'too_many_values' validation error, unless 'filters' includes a value of its parent type.
  -h, --help                              help for attributed-usage
      --limit string                      Results per page (1-1000)
      --page string                       Page number. Defaults to 1. A value that is not a positive integer, such as '0', returns the first page.
      --per-page string                   Number of attribution values per page, between 1 and 100.
      --search-term string                Filter the attribution values by search term, up to 100 characters. The totals are not affected by the search.
      --split-charge-code string          Code of a charge of the plan to split by charge filter, with one entry per filter in 'charges_usage'. The filters are returned 10 at a time, in the order of the charge, starting at 'split_filters_offset'. The charge must have at least 1 filter, and must be part of 'charge_codes[]' when it is sent. A 'too_many_filters' validation error is returned when the filter values of the charge are too many to match events.
      --split-filters-offset string       Index of the first filter of the split charge to return, starting at 0. Use it with 'split_filters_count' to page through the filters 10 at a time. It is ignored when 'split_charge_code' is not sent.
      --to-datetime string                End of the period, in ISO 8601 format. It must be sent together with 'from_datetime'.
      --watch                             Poll and re-render when the response changes
      --watch-interval duration           Polling interval used with --watch (default 2s)
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
