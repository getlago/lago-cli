## lago v2 credit-notes create

Create credit note

### Synopsis

Requires a premium license and an eligible invoice. Credits are allocated across credit, refund and offset amounts. Product and legacy invoice fees can be credited.
Requires the product catalog to be enabled and credit_note write permission on the API key.

```
lago v2 credit-notes create [flags]
```

### Examples

```
  lago v2 credit-notes create --input @payload.json
  lago v2 credit-notes create --input @payload.json --output json  # full resource
```

### Options

```
      --credit-amount-cents string   Amount credited to the customer balance, in cents.
      --description string           Description of the adjustment.
  -h, --help                         help for create
      --input string                 Complete JSON request body or @file.json
      --invoice-id string            Invoice identifier.
      --items string                 Invoice fees to credit.
      --metadata string              Metadata as an object of key-value pairs. At most 50 keys are allowed. Keys are limited to 100 characters and values to 255 characters. Values must be strings or null.
      --offset-amount-cents string   Amount offset against the invoice balance, in cents.
      --reason string                Reason for issuing the credit note.; one of: duplicated_charge, product_unsatisfactory, order_change, order_cancellation, fraudulent_charge, other
      --refund-amount-cents string   Amount to refund, in cents.
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

* [lago v2 credit-notes](lago_v2_credit-notes)	 - Manage Lago credit notes
