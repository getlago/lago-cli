## lago customers list

List all customers

### Synopsis

This endpoint retrieves all existing customers.

```
lago customers list [flags]
```

### Examples

```
  lago customers list
```

### Options

```
      --account-type string                    Filter customers by account type.
      --all                                    Fetch every page
      --billing-entity-codes string            Filter customers by billing entity codes.
      --countries string                       Filter customers by countries. Possible values are the ISO 3166-1 alpha-2 codes.
      --currencies string                      Filter customers by currencies.
      --customer-type string                   Filter customers by customer type.; one of: company, individual
      --external-id string                     Filter customers by external ID.
      --has-customer-type string               Filter customers by whether they have a customer type or not.
      --has-tax-identification-number string   Filter customers by whether they have a tax identification number or not.
  -h, --help                                   help for list
      --limit string                           Maximum number of results
      --metadata[key] string                   Filter customers by metadata. Replace 'key' with the actual metadata key you want to match, and provide the corresponding value. Providing empty value will search for customers without given metadata key. For example, 'metadata[is_synced]=true&metadata[last_synced_at]='.
      --page string                            Page number.
      --per-page string                        Number of records per page.
      --search-term string                     Filter customers by search term. This will filter all customers whose name, firstname, lastname, legal name, external id or email contain the search term.
      --states string                          Filter customers by states.
      --watch                                  Poll and re-render when the response changes
      --watch-interval duration                Polling interval used with --watch (default 2s)
      --zipcodes string                        Filter customers by zipcodes.
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
