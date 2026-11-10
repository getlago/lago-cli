## lago v2 contracts create

Create contract

### Synopsis

Creates a contract for an existing customer. The plan is optional. One pending or active contract may use an external_id at a time. A plan supplies the initial applied rate cards.
Requires the product catalog to be enabled and contract write permission on the API key.

```
lago v2 contracts create [flags]
```

### Examples

```
  lago v2 contracts create --input @payload.json
  lago v2 contracts create --input @payload.json --output json  # full resource
```

### Options

```
      --billing-anchor-date string                                   Optional billing anchor.
      --billing-time string                                          Whether billing aligns to the calendar or contract anniversary.; one of: calendar, anniversary
      --consolidate-invoice string                                   Whether invoices can consolidate compatible contracts.
      --ended-at string                                              Optional end timestamp. Must be after the start and not already in the past.
      --external-customer-id string                                  External customer identifier.
      --external-id string                                           Unique identifier for a live contract in your application.
  -h, --help                                                         help for create
      --input string                                                 Complete JSON request body or @file.json
      --invoice-custom-section-invoice-custom-section-codes string   Invoice custom section codes.
      --invoice-custom-section-skip-invoice-custom-sections string   Whether to skip invoice custom sections.
      --name string                                                  Contract name.
      --plan-code string                                             Optional plan code. Plan-less contracts use directly attached rate cards.
      --purchase-order-number string                                 Purchase order reference.
      --started-at string                                            Start timestamp. Defaults to now. A future start creates a pending contract.
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

* [lago v2 contracts](lago_v2_contracts)	 - Manage Lago contracts
