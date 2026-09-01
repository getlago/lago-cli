## lago wallets create-wallet-transaction

Top up a wallet

### Synopsis

This endpoint is used to top-up an active wallet.

```
lago wallets create-wallet-transaction [flags]
```

### Examples

```
  lago wallets create-wallet-transaction --input @payload.json
  lago wallets create-wallet-transaction --input @payload.json --output json  # full resource
```

### Options

```
      --granted-credits string                                       The number of free granted credits.
  -h, --help                                                         help for create-wallet-transaction
      --idempotency-key string                                       Idempotency key for safe mutation retries
      --ignore-paid-top-up-limits string                             When true, allows topping up the wallet with transactions that exceed the paid top-up limits. Defaults to false.
      --input string                                                 Complete JSON request body or @file.json
      --invoice-custom-section-invoice-custom-section-codes string   List of unique codes identifying the invoice custom sections to apply. These override the default invoice custom sections configured at the customer or billing entity level.
      --invoice-custom-section-skip-invoice-custom-sections string   Set to 'true' to exclude all invoice custom sections from PDF generation for invoices related to this resource. When 'true', 'invoice_custom_section_codes' is ignored.
      --invoice-requires-successful-payment string                   A boolean setting that, when set to true, delays issuing an invoice for a wallet top-up until a successful payment is made; if false, the invoice is issued immediately upon wallet top-up, regardless of the payment status. Default value of false.
      --metadata string                                              This optional field allows you to store a list of key-value pairs that hold additional information or custom attributes related to the data.
      --name string                                                  The name of the wallet transaction. It will appear on the invoice as the label for the fee. If not set, the label on the invoice will fallback to either 'Prepaid credits - {{wallet_name}}' if the wallet name is set, or 'Prepaid credits'.

                                                                     Note that this name will apply to all transactions ('paid_credits', 'granted_credits' and 'voided_credits') created by this action.
      --paid-credits string                                          The number of paid credits.
      --payment-method-payment-method-id string                      The unique identifier of the payment method (required when using a specific provider payment method).
      --payment-method-payment-method-type string                    The type of payment method to use.; one of: provider, manual
      --purchase-order-number string                                 The purchase order number associated with this wallet transaction. It will be added to invoices generated for the resulting wallet top-up. If not set, falls back to the wallet's 'purchase_order_number'.
      --voided-credits string                                        The number of voided credits.
      --wallet-id string                                             Unique identifier assigned to the wallet within the Lago application. This ID is exclusively created by Lago and serves as a unique identifier for the wallet's record within the Lago system.
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

* [lago wallets](lago_wallets)	 - Manage Lago wallets
