## lago v2 taxes delete

Delete a tax

### Synopsis

Deletes a tax identified by its code and returns it with a non-null deleted_at timestamp.
Requires the product catalog to be enabled and tax write permission on the API key.
Deleted taxes are no longer returned by listing or retrieval. An unknown or already deleted tax returns 404.
Taxes have no supported expansions; a nonempty expands[] returns 400 before deletion.

```
lago v2 taxes delete <code> [flags]
```

### Examples

```
  lago v2 taxes delete <code>
  lago v2 taxes delete <code> --output json  # full resource
```

### Options

```
  -h, --help   help for delete
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
