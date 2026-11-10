## lago v2 credit-notes delete-metadata-key

Delete one credit note metadata key

### Synopsis

Deletes one metadata key from a finalized credit note and returns the remaining metadata.
Requires the product catalog to be enabled and credit_note write permission on the API key.

```
lago v2 credit-notes delete-metadata-key <credit_note_id> <key> [flags]
```

### Examples

```
  lago v2 credit-notes delete-metadata-key <credit_note_id> <key>
  lago v2 credit-notes delete-metadata-key <credit_note_id> <key> --output json  # full resource
```

### Options

```
  -h, --help   help for delete-metadata-key
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

* [lago v2 credit-notes](lago_v2_credit-notes)	 - Manage Lago credit notes
