## lago v2 taxes patch-tax

Partially update a tax

### Synopsis

Updates only the supplied fields of a non-deleted tax identified by its code.
Requires the product catalog to be enabled and tax write permission on the API key.
Behaves the same as PUT. The tax wrapper must be nonempty, and omitted fields keep their current values.
The applied_to_organization parameter is ignored and existing assignments are not changed by this parameter.
Taxes have no supported expansions; a nonempty expands[] returns 400 before updating the tax.

```
lago v2 taxes patch-tax <code> [flags]
```

### Examples

```
  lago v2 taxes patch-tax <code> --input @payload.json
  lago v2 taxes patch-tax <code> --input @payload.json --output json  # full resource
```

### Options

```
      --code string          Unique code of the tax within the organization, including deleted taxes. It cannot be blank.
      --description string   Internal description of the tax. Set to null to clear it.
  -h, --help                 help for patch-tax
      --input string         Complete JSON request body or @file.json
      --name string          Name of the tax. It cannot be blank.
      --rate string          Percentage rate of the tax. Accepts a number or a decimal string.
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

* [lago v2 taxes](lago_v2_taxes)	 - Manage Lago taxes
