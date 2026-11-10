## lago v2 customers set-as-default

Set the default customer payment method

### Synopsis

Sets this payment method as the customer default. No request body is required.
Requires the product catalog to be enabled and payment_method write permission on the API key.

```
lago v2 customers set-as-default <customer_external_id> <lago_id> [flags]
```

### Examples

```
  lago v2 customers set-as-default <customer_external_id> <lago_id>
  lago v2 customers set-as-default <customer_external_id> <lago_id> --output json  # full resource
```

### Options

```
  -h, --help   help for set-as-default
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

* [lago v2 customers](lago_v2_customers)	 - Manage Lago customers
