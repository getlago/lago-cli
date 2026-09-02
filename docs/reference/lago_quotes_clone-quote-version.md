## lago quotes clone-quote-version

Clone a quote version

### Synopsis

This endpoint copies a quote version into a new `draft` version of the same quote, so that the deal can be renegotiated.
Any version of the quote can be cloned, but a quote can only carry one active version: the draft the quote currently holds, if any, is voided with the `superseded` reason. Cloning is therefore rejected with a `422` once a version of the quote has been approved.
A concurrent write on the same quote is reported as a `422` rather than retried, so a duplicated call can fail while the first one is still in flight.
This is a premium feature.

```
lago quotes clone-quote-version <lago_id> [flags]
```

### Examples

```
  lago quotes clone-quote-version <lago_id>
```

### Options

```
  -h, --help   help for clone-quote-version
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
