## lago payment-methods

Manage Lago payment methods

### Options

```
  -h, --help   help for payment-methods
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
* [lago payment-methods destroy](lago_payment-methods_destroy)	 - Delete a payment method
* [lago payment-methods list-customers](lago_payment-methods_list-customers)	 - List all customer payment methods
* [lago payment-methods payment-method-set-as-default](lago_payment-methods_payment-method-set-as-default)	 - Set the payment method as default
