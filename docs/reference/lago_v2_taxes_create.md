## lago v2 taxes create

Create a tax

### Synopsis

Creates a tax rate for the organization.
Requires the product catalog to be enabled and tax write permission on the API key.
Name, code and rate are required. A code already used by another tax of the organization, including a deleted one, returns 422.
The applied_to_organization parameter is ignored and is not returned. Creating a tax does not assign it to any resource.
Taxes have no supported expansions; a nonempty expands[] returns 400 before creation.

```
lago v2 taxes create [flags]
```

### Examples

```
  lago v2 taxes create --input @payload.json
  lago v2 taxes create --input @payload.json --output json  # full resource
```

### Options

```
      --code string          Unique code of the tax within the organization, including deleted taxes. It cannot be blank.
      --description string   Internal description of the tax. Set to null to clear it.
  -h, --help                 help for create
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
