## lago v2 customers destroy-customer-payment-method

Delete a customer payment method

### Synopsis

Deletes the payment method and returns it with deleted_at set.
Requires the product catalog to be enabled and payment_method write permission on the API key.

```
lago v2 customers destroy-customer-payment-method <customer_external_id> <lago_id> [flags]
```

### Examples

```
  lago v2 customers destroy-customer-payment-method <customer_external_id> <lago_id>
  lago v2 customers destroy-customer-payment-method <customer_external_id> <lago_id> --output json  # full resource
```

### Options

```
  -h, --help   help for destroy-customer-payment-method
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
