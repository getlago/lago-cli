## lago v2 credit-notes update

Update credit note

### Synopsis

PUT and PATCH both accept partial updates. Omitted fields are unchanged.
Only refund_status and metadata are accepted. Metadata updates merge keys. Draft credit notes return 404.
Requires the product catalog to be enabled and credit_note write permission on the API key.

```
lago v2 credit-notes update <lago_id> [flags]
```

### Examples

```
  lago v2 credit-notes update <lago_id> --input @payload.json
  lago v2 credit-notes update <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help                   help for update
      --input string           Complete JSON request body or @file.json
      --metadata string        Metadata as an object of key-value pairs. At most 50 keys are allowed. Keys are limited to 100 characters and values to 255 characters. Values must be strings or null.
      --refund-status string   Refund status.; one of: pending, succeeded, failed,
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
