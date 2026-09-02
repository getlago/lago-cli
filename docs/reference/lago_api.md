## lago api

Make an authenticated request to any Lago API endpoint

```
lago api METHOD PATH [flags]
```

### Examples

```
  lago api GET /customers?page=2
  lago api POST /events --data @event.json
  printf '{"event":{}}' | lago api POST /events --data -
```

### Options

```
  -d, --data string      Request body JSON, @file, or - for stdin
  -H, --header strings   Additional request header (repeatable)
  -h, --help             help for api
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

* [lago](lago)	 - The official CLI for Lago billing
