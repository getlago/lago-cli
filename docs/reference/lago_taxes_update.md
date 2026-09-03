## lago taxes update

Update a tax

### Synopsis

This endpoint updates an existing tax representing a customizable tax rate applicable to either the organization or a specific customer.

```
lago taxes update <code> [flags]
```

### Examples

```
  lago taxes update <code> --input @payload.json
  lago taxes update <code> --input @payload.json --output json  # full resource
```

### Options

```
      --applied-to-organization string   **Deprecated.** This field will be removed in a future version. When set to true, it applies the tax to the organization's default billing entity. To apply or remove a tax from any billing entity (including the default one), please use the 'PUT /billing_entities/:code' endpoint instead.
      --code string                      Unique code used to identify the tax associated with the API request.
      --description string               Internal description of the tax
  -h, --help                             help for update
      --input string                     Complete JSON request body or @file.json
      --name string                      Name of the tax.
      --rate string                      The percentage rate of the tax
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

* [lago taxes](lago_taxes)	 - Manage Lago taxes
