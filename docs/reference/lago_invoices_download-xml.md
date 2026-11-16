## lago invoices download-xml

Download an invoice XML

### Synopsis

This endpoint is used for downloading the e-invoicing XML document of a specific invoice. Only `finalized` invoices can be downloaded; any other invoice returns a `404` error.
When the XML has already been generated, the invoice object is returned and its `xml_url` points to the document. When it has not been generated yet, the response is a `200` with an empty body and the generation starts in the background. Call the endpoint again to get the invoice object with its `xml_url`.

```
lago invoices download-xml <lago_id> [flags]
```

### Examples

```
  lago invoices download-xml <lago_id>
  lago invoices download-xml <lago_id> --output json  # full resource
```

### Options

```
  -h, --help   help for download-xml
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

* [lago invoices](lago_invoices)	 - Manage Lago invoices
