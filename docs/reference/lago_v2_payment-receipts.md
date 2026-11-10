## lago v2 payment-receipts

Manage Lago payment receipts

### Options

```
  -h, --help   help for payment-receipts
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

* [lago v2](lago_v2)	 - Lago API v2 (beta): product catalog, contracts and v2 resources
* [lago v2 payment-receipts get](lago_v2_payment-receipts_get)	 - Retrieve payment receipt
* [lago v2 payment-receipts list](lago_v2_payment-receipts_list)	 - List payment receipts
* [lago v2 payment-receipts resend-email](lago_v2_payment-receipts_resend-email)	 - Resend a payment receipt email
