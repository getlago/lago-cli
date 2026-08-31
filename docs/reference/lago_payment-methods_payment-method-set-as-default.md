## lago payment-methods payment-method-set-as-default

Set the payment method as default

### Synopsis

Use the payment method as default when not selected a payment method

```
lago payment-methods payment-method-set-as-default <lago_id> <external_customer_id> [flags]
```

### Examples

```
  lago payment-methods payment-method-set-as-default <lago_id> <external_customer_id>
```

### Options

```
  -h, --help   help for payment-method-set-as-default
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
