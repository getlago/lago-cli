## lago events send

Send usage events

### Synopsis

This endpoint is used for transmitting usage measurement events to either a designated customer or a specific subscription.

```
lago events send [flags]
```

### Examples

```
  lago events send --input @payload.json
```

### Options

```
      --code string                         The code that identifies a targeted billable metric. It is essential that this code matches the 'code' property of one of your active billable metrics. If the provided code does not correspond to any active billable metric, it will be ignored during the process.
      --concurrency int                     Concurrent bulk event requests (1-64) (default 4)
      --external-subscription-id string     The unique identifier of the subscription in your application. This field is mandatory in order to link events to the correct customer subscription.
      --file string                         Stream newline-delimited JSON events from a file, or - for stdin
  -h, --help                                help for send
      --input string                        Complete JSON request body or @file.json
      --precise-total-amount-cents string   The precise total amount in cents with precision used by the 'dynamic' pricing model to compute the usage amount.
      --properties string                   This field represents additional properties associated with the event, which are utilized in the calculation of the final fee. This object becomes mandatory when the targeted billable metric employs a 'sum_agg', 'max_agg', or 'unique_count_agg' aggregation method. However, when using a simple 'count_agg', this object is not required.
      --timestamp string                    This field captures the Unix timestamp in seconds indicating the occurrence of the event in Coordinated Universal Time (UTC).
                                            If this timestamp is not provided, the API will automatically set it to the time of event reception.
                                            You can also provide miliseconds precision by appending decimals to the timestamp.
      --transaction-id string               This field represents a unique identifier for the event.
                                            It is crucial for ensuring idempotency, meaning that each event can be uniquely identified and processed without causing any unintended side effects.

                                            WARNING: If the Lago organization is configured to use the new Clickhouse-based event pipeline (designed for high-volume processing), the idempotency logic is handled differently.
                                            Event uniqueness is maintained with both 'transaction_id' and 'timestamp' fields.
                                            If a new event arrives with identical values for these two fields as an existing event, the new one will overwrite the previous event rather than being rejected.
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
