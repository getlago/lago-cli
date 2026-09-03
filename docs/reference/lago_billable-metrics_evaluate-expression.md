## lago billable-metrics evaluate-expression

Evaluate an expression for a billable metric

### Synopsis

Evaluate an expression for a billable metric creation by providing the expression and test data

```
lago billable-metrics evaluate-expression [flags]
```

### Examples

```
  lago billable-metrics evaluate-expression --input @payload.json
```

### Options

```
      --code string         The code that identifies a targeted billable metric.
      --expression string   Expression used to calculate the event units. The expression is evalutated for each event and the result is then used to calculate the total aggregated units.
                            Accepted function are 'ceil', 'concat' and 'round' as well as '+', '-', '\' and '*' operations.
                            Round is accepting an optional second parameter to specify the number of decimal.
  -h, --help                help for evaluate-expression
      --input string        Complete JSON request body or @file.json
      --properties string   This field represents additional properties associated with the event. They can be used when evaluating the expression.
      --timestamp string    This field captures the Unix timestamp in seconds indicating the occurrence of the event in Coordinated Universal Time (UTC).
                            If this timestamp is not provided, the API will automatically set it to the time of event reception.
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

* [lago billable-metrics](lago_billable-metrics)	 - Manage Lago billable metrics
