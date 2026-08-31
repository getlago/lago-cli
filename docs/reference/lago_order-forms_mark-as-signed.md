## lago order-forms mark-as-signed

Mark an order form as signed

### Synopsis

This endpoint records the customer's signature on an order form, and creates the order that carries the quoted deal out.
Only a `generated` order form can be signed, so a `422` is returned once the order form has been signed, has expired or has been voided.
An `execute_at` landing on or after the day the quoted deal stops being executable is rejected as well, so the order can never be scheduled past the term it bills.
A concurrent write on the same quote is reported as a `422` rather than retried, so a duplicated call can fail while the first one is still in flight.
This is a premium feature.

```
lago order-forms mark-as-signed <lago_id> [flags]
```

### Examples

```
  lago order-forms mark-as-signed <lago_id> --input @payload.json
```

### Options

```
      --execute-at string        The date and time in UTC (ISO 8601) when Lago executes the order created from the order form. It must be in the future, requires 'execution_mode' to be set, and must fall strictly before the day the quoted deal stops being executable, otherwise signing is rejected with a '422'. When it is omitted, the order is created but not scheduled, and waits to be executed on demand.
      --execution-mode string    How the order created from the order form is carried out. It becomes mandatory as soon as 'execute_at' is provided.; one of: execute_in_lago, order_only
  -h, --help                     help for mark-as-signed
      --idempotency-key string   Idempotency key for safe mutation retries
      --input string             Complete JSON request body or @file.json
      --signed-document string   The document signed by the customer, as a base64 data URI ('data:<content_type>;base64,<data>'). Accepted content types are 'application/pdf', 'image/jpeg' and 'image/png', for a maximum of 10 MB. Once attached, it is exposed as 'signed_document_url' on the order form.
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
