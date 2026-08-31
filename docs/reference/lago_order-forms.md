## lago order-forms

Manage Lago order forms

### Options

```
  -h, --help   help for order-forms
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
* [lago order-forms get](lago_order-forms_get)	 - Retrieve an order form
* [lago order-forms list](lago_order-forms_list)	 - List all order forms
* [lago order-forms mark-as-signed](lago_order-forms_mark-as-signed)	 - Mark an order form as signed
* [lago order-forms void](lago_order-forms_void)	 - Void an order form
