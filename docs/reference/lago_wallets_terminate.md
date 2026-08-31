## lago wallets terminate

Terminate a wallet

### Synopsis

This endpoint is used to terminate an existing wallet with prepaid credits.

```
lago wallets terminate <lago_id> [flags]
```

### Examples

```
  lago wallets terminate <lago_id>
```

### Options

```
  -h, --help                     help for terminate
      --idempotency-key string   Idempotency key for safe mutation retries
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
