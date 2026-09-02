## lago orders execute

Execute an order

### Synopsis

This endpoint carries out an order on demand, without waiting for its `execute_at` schedule. It runs synchronously, so the response carries the resulting status and, when the execution fails, the reason it failed.
An `execute_in_lago` order applies the quoted deal, creating the subscriptions, coupons, wallets or one-off invoice it describes; an `order_only` order is simply marked as executed. What the execution produced is listed in `execution_record`.
Executing is idempotent: an order that has already been executed is returned untouched. An order whose previous execution failed is executed again, and a `422` is returned when the order carries no execution mode and none is provided in the body.
Restating an `execution_mode` that differs from the one stored on the order updates it first, which only succeeds while the order is still `created`: a `failed` order can be retried, but not switched to another mode. That update also re-validates the order's own `execute_at` against the day the quoted deal stops being executable, so a `422` is possible even though no date was sent. An order carrying no `execute_at` is not bounded that way, and a deal whose term has passed then only fails when the execution runs.
A `404` is not limited to an unknown order: when the execution cannot find a catalog record the quote pinned, such as a deleted plan, charge, coupon or the subscription being amended, the missing resource is reported as a `404` and the order is left `failed`.
Concurrency is reported two ways. A mode change losing the race on the quote comes back as a `422` validation error, while the execution itself losing it comes back as a `422` carrying a lock code, because a scheduled execution relies on that error to be retried.
This is a premium feature.

```
lago orders execute <lago_id> [flags]
```

### Examples

```
  lago orders execute <lago_id> --input @payload.json
  lago orders execute <lago_id> --input @payload.json --output json  # full resource
```

### Options

```
      --execution-mode string   How the order is carried out. It is only persisted when it differs from the mode already stored on the order, so restating the current mode never fails. Changing the mode of an order that has already been executed is rejected.; one of: execute_in_lago, order_only
  -h, --help                    help for execute
      --input string            Complete JSON request body or @file.json
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

* [lago orders](lago_orders)	 - Manage Lago orders
