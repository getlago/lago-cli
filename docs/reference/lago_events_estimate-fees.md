## lago events estimate-fees

Estimate fees for a pay in advance charge

### Synopsis

Estimate the fees that would be created after reception of an event for a billable metric attached to one or multiple pay in advance charges

```
lago events estimate-fees [flags]
```

### Examples

```
  lago events estimate-fees --input @payload.json
```

### Options

```
      --code string                       The code that identifies a targeted billable metric. It is essential that this code matches the 'code' property of one of your active billable metrics. If the provided code does not correspond to any active billable metric, it will be ignored during the process.
      --external-subscription-id string   The unique identifier of the subscription within your application.
  -h, --help                              help for estimate-fees
      --input string                      Complete JSON request body or @file.json
      --properties string                 This field represents additional properties associated with the event, which are utilized in the calculation of the final fee. This object becomes mandatory when the targeted billable metric employs a 'sum_agg', 'max_agg', or 'unique_count_agg' aggregation method. However, when using a simple 'count_agg', this object is not required. Any property of the targeted billable metric can be sent here, in addition to the documented fields below.
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

* [lago events](lago_events)	 - Manage Lago events
