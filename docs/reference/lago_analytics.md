## lago analytics

Manage Lago analytics

### Options

```
  -h, --help   help for analytics
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
* [lago analytics gross-revenue](lago_analytics_gross-revenue)	 - List gross revenue
* [lago analytics invoice-collection](lago_analytics_invoice-collection)	 - List of finalized invoices
* [lago analytics invoiced-usage](lago_analytics_invoiced-usage)	 - List usage revenue
* [lago analytics mrr](lago_analytics_mrr)	 - List MRR
* [lago analytics overdue-balance](lago_analytics_overdue-balance)	 - List overdue balance
* [lago analytics usage](lago_analytics_usage)	 - List usage
