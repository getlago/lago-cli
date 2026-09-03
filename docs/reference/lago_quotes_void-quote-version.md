## lago quotes void-quote-version

Void a quote version

### Synopsis

This endpoint voids a quote version, which makes it definitive: a voided version can neither be edited nor approved again.
Only a `draft` version can be voided through this endpoint, and the resulting `void_reason` is `manual`. An `approved` version is voided by Lago itself, through the order form generated from it.
A concurrent write on the same quote is reported as a `422` rather than retried, so a duplicated call can fail while the first one is still in flight.
This is a premium feature.

```
lago quotes void-quote-version <lago_id> [flags]
```

### Examples

```
  lago quotes void-quote-version <lago_id>
  lago quotes void-quote-version <lago_id> --output json  # full resource
```

### Options

```
  -h, --help   help for void-quote-version
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

* [lago quotes](lago_quotes)	 - Manage Lago quotes
