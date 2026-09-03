## lago payment-methods destroy

Delete a payment method

### Synopsis

This endpoint deletes a specific payment method for a customer.

```
lago payment-methods destroy <lago_id> <external_customer_id> [flags]
```

### Examples

```
  lago payment-methods destroy <lago_id> <external_customer_id>
  lago payment-methods destroy <lago_id> <external_customer_id> --output json  # full resource
```

### Options

```
  -h, --help   help for destroy
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

* [lago payment-methods](lago_payment-methods)	 - Manage Lago payment methods
