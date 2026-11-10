## lago v2 taxes get

Retrieve a tax

### Synopsis

Retrieves a non-deleted tax by its code.
Requires the product catalog to be enabled and tax read permission on the API key.
Unknown taxes and taxes belonging to another organization return 404.
Assignments and relationship counts are not returned. Taxes have no supported expansions; a nonempty expands[] returns 400.

```
lago v2 taxes get <code> [flags]
```

### Examples

```
  lago v2 taxes get <code>
```

### Options

```
  -h, --help                      help for get
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

* [lago v2 taxes](lago_v2_taxes)	 - Manage Lago taxes
