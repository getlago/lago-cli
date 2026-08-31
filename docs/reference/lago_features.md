## lago features

Manage Lago features

### Options

```
  -h, --help   help for features
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
* [lago features create](lago_features_create)	 - Create a feature
* [lago features delete](lago_features_delete)	 - Delete a feature
* [lago features delete-feature-privilege](lago_features_delete-feature-privilege)	 - Delete a privilege. Deleting a privilege removes it from all plans and subscriptions.
* [lago features get](lago_features_get)	 - Retrieve a feature
* [lago features list](lago_features_list)	 - List all features
* [lago features update](lago_features_update)	 - Update a feature
