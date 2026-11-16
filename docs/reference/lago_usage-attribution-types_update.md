## lago usage-attribution-types update

Update a usage attribution type

### Synopsis

This endpoint updates an existing usage attribution type. Only the fields present in the payload are updated. Once usage is attributed to the type, only its `name`, `description` and `attribution_keys` can be changed.

```
lago usage-attribution-types update <code> [flags]
```

### Examples

```
  lago usage-attribution-types update <code> --input @payload.json
  lago usage-attribution-types update <code> --input @payload.json --output json  # full resource
```

### Options

```
      --attribution-keys string   Keys of the event properties that attribute usage to this type, up to 4. Blank and duplicate keys are removed. A key already used by another usage attribution type of the organization is rejected.
      --code string               Unique code used to identify the usage attribution type within the organization. Once usage is attributed to the type, the code cannot be changed.
      --description string        Description of the usage attribution type.
  -h, --help                      help for update
      --input string              Complete JSON request body or @file.json
      --name string               Name of the usage attribution type.
      --parent-code string        Code of the parent usage attribution type, which must be 'hierarchical'. Send 'null' or an empty string to remove the parent. Once usage is attributed to the type, the parent cannot be changed.
      --role string               Role of the usage attribution type. A 'flat' type cannot have a parent, and a type with children cannot become 'flat'. Once usage is attributed to the type, the role cannot be changed.; one of: hierarchical, flat
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

* [lago usage-attribution-types](lago_usage-attribution-types)	 - Manage Lago usage attribution types
