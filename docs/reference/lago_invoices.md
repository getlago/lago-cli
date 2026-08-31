## lago invoices

Manage Lago invoices

### Options

```
  -h, --help   help for invoices
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
* [lago invoices create](lago_invoices_create)	 - Create a one-off invoice
* [lago invoices delete](lago_invoices_delete)	 - Delete a draft invoice
* [lago invoices download](lago_invoices_download)	 - Download an invoice PDF
* [lago invoices finalize](lago_invoices_finalize)	 - Finalize a draft invoice
* [lago invoices get](lago_invoices_get)	 - Retrieve an invoice
* [lago invoices list](lago_invoices_list)	 - List all invoices
* [lago invoices lose-dispute](lago_invoices_lose-dispute)	 - Mark an invoice payment dispute as lost
* [lago invoices payment-url](lago_invoices_payment-url)	 - Generate a payment URL
* [lago invoices preview](lago_invoices_preview)	 - Create an invoice preview
* [lago invoices refresh](lago_invoices_refresh)	 - Refresh a draft invoice
* [lago invoices retry](lago_invoices_retry)	 - Retry generation of a failed invoice
* [lago invoices retry-payment](lago_invoices_retry-payment)	 - Retry an invoice payment
* [lago invoices update](lago_invoices_update)	 - Update an invoice
* [lago invoices void](lago_invoices_void)	 - Void an invoice
