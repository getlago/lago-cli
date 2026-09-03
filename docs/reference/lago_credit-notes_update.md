## lago credit-notes update

Update a credit note

### Synopsis

This endpoint updates an existing credit note.

```
lago credit-notes update <lago_id> [flags]
```

### Examples

```
  lago credit-notes update <lago_id> --input @payload.json
  lago credit-notes update <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help                   help for update
      --input string           Complete JSON request body or @file.json
      --metadata string        Metadata to set as key-value pairs. Keys are strings (max 100 characters), values can be strings (max 255 characters) or null.
      --refund-status string   The status of the refund portion of the credit note. It indicates the current state or condition of the refund associated with the credit note. The possible values for this field are:

                               - 'pending': this status indicates that the refund is pending execution. The refund request has been initiated but has not been processed or completed yet.
                               - 'succeeded': this status indicates that the refund has been successfully executed. The refund amount has been processed and returned to the customer or the designated recipient.
                               - 'failed': this status indicates that the refund failed to execute. The refund request encountered an error or unsuccessful processing, and the refund amount could not be returned.; one of: pending, succeeded, failed
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

* [lago credit-notes](lago_credit-notes)	 - Manage Lago credit notes
