## lago features update

Update a feature

### Synopsis

This endpoint updates an existing feature representing an entitlement component of your application.

```
lago features update <code> [flags]
```

### Examples

```
  lago features update <code> --input @payload.json
  lago features update <code> --input @payload.json --output json  # full resource
```

### Options

```
      --description string   Internal description of the feature.
  -h, --help                 help for update
      --input string         Complete JSON request body or @file.json
      --name string          Name of the feature.
      --privileges string    Privileges associated with this feature. Can be empty.
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

* [lago features](lago_features)	 - Manage Lago features
