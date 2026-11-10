## lago v2 organizations update

Update the organization

### Synopsis

Updates email and slug. Omitted fields are unchanged. Name and events_store are read-only.
Requires the product catalog to be enabled and organization write permission on the API key.

```
lago v2 organizations update [flags]
```

### Examples

```
  lago v2 organizations update --input @payload.json
  lago v2 organizations update --input @payload.json --output json  # full resource
```

### Options

```
      --email string   Organization email.
  -h, --help           help for update
      --input string   Complete JSON request body or @file.json
      --slug string    Organization slug.
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

* [lago v2 organizations](lago_v2_organizations)	 - Manage Lago organizations
