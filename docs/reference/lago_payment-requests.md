## lago payment-requests

Manage Lago payment requests

### Options

```
  -h, --help   help for payment-requests
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
* [lago payment-requests create](lago_payment-requests_create)	 - Create a payment request
* [lago payment-requests get](lago_payment-requests_get)	 - Retrieve a payment request
* [lago payment-requests list](lago_payment-requests_list)	 - List all payment requests
