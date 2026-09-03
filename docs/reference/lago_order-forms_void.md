## lago order-forms void

Void an order form

### Synopsis

This endpoint voids an order form, with the `manual` reason, and cascades to the quote version it was generated from, which is voided with the `cascade_of_voided` reason.
Only a `generated` order form can be voided, so a `422` is returned once the order form has been signed, has expired or has already been voided.
A concurrent write on the same quote is reported as a `422` rather than retried, so a duplicated call can fail while the first one is still in flight.
This is a premium feature.

```
lago order-forms void <lago_id> [flags]
```

### Examples

```
  lago order-forms void <lago_id>
```

### Options

```
  -h, --help   help for void
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

* [lago order-forms](lago_order-forms)	 - Manage Lago order forms
