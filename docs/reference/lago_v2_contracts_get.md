## lago v2 contracts get

Retrieve contract

### Synopsis

Without status, retrieves the live pending or active contract. A status can select a historical terminated or canceled contract.
Requires the product catalog to be enabled and contract read permission on the API key.

```
lago v2 contracts get <external_id> [flags]
```

### Examples

```
  lago v2 contracts get <external_id>
```

### Options

```
      --expands string            Relationships to include in this response. Only one expansion level is supported.
  -h, --help                      help for get
      --status string             Selects a lifecycle status. An unknown value falls back to active.; one of: pending, active, terminated, canceled
      --watch                     Poll and re-render when the response changes
      --watch-interval duration   Polling interval used with --watch (default 2s)
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
