## lago wallets merge-wallet-metadata

Merge wallet metadata

### Synopsis

This endpoint merges the provided metadata with existing metadata on the wallet.
Existing keys not in the request are preserved. New keys are added, existing keys are updated.

```
lago wallets merge-wallet-metadata <lago_id> [flags]
```

### Examples

```
  lago wallets merge-wallet-metadata <lago_id> --input @payload.json
  lago wallets merge-wallet-metadata <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
  -h, --help           help for merge-wallet-metadata
      --input string   Complete JSON request body or @file.json
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

* [lago wallets](lago_wallets)	 - Manage Lago wallets
