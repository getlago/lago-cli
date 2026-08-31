## lago wallets update

Update a wallet

### Synopsis

This endpoint is used to update an existing wallet with prepaid credits. A terminated wallet cannot be updated and returns a `422` error.

```
lago wallets update <lago_id> [flags]
```

### Examples

```
  lago wallets update <lago_id> --input @payload.json
```

### Options

```
      --applies-to-billable-metric-codes string                      An array of billable metric codes to which the wallet is applicable. By specifying the billable metric codes in this field, you can restrict the wallet's usage to specific metrics only.
      --applies-to-fee-types string                                  An array of fee types to which the wallet is applicable. By specifying the fee types in this field, you can restrict the wallet's usage to specific fee types only.
      --billing-entity-code string                                   The code of the billing entity associated with the wallet.
      --code string                                                  The code of the wallet.
      --expiration-at string                                         The date and time that determines when the wallet will expire. It follows the ISO 8601 datetime format and is expressed in Coordinated Universal Time (UTC).
  -h, --help                                                         help for update
      --input string                                                 Complete JSON request body or @file.json
      --invoice-custom-section-invoice-custom-section-codes string   List of unique codes identifying the invoice custom sections to apply. These override the default invoice custom sections configured at the customer or billing entity level.
      --invoice-custom-section-skip-invoice-custom-sections string   Set to 'true' to exclude all invoice custom sections from PDF generation for invoices related to this resource. When 'true', 'invoice_custom_section_codes' is ignored.
      --invoice-requires-successful-payment string                   A boolean setting that, when set to true, delays issuing an invoice for a wallet top-up until a successful payment is made; if false, the invoice is issued immediately upon wallet top-up, regardless of the payment status. Default value of false.
      --metadata string                                              Metadata to set as key-value pairs. Keys are strings (max 100 characters), values can be strings (max 255 characters) or null.
      --name string                                                  The name of the wallet.
      --payment-method-payment-method-id string                      The unique identifier of the payment method (required when using a specific provider payment method).
      --payment-method-payment-method-type string                    The type of payment method to use.; one of: provider, manual
      --priority string                                              Wallet priority for ordering when a customer has multiple wallets. Allowed values: 1-50, where 1 is highest priority and 50 is lowest. Defaults to 50.
      --purchase-order-number string                                 The purchase order number associated with the wallet. It will be added to invoices generated for wallet top-ups, unless a more specific purchase order number is set on the wallet transaction or recurring transaction rule that triggered the top-up.
      --recurring-transaction-rules string                           List of recurring transaction rules. Currently, we only allow one recurring rule per wallet.
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
