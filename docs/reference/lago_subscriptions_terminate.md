## lago subscriptions terminate

Terminate a subscription

### Synopsis

This endpoint allows you to terminate a subscription.

```
lago subscriptions terminate <external_id> [flags]
```

### Examples

```
  lago subscriptions terminate <external_id>
```

### Options

```
  -h, --help                                help for terminate
      --idempotency-key string              Idempotency key for safe mutation retries
      --on-termination-credit-note string   When a pay-in-advance subscription is terminated before the end of its billing period, we generate a credit note for the unused subscription time by default.
                                            This field allows you to control the behavior of the credit note generation:

                                            - 'credit': A credit note is generated for the unused subscription time. The unused amount is credited back to the customer.
                                            - 'refund': A credit note is generated for the unused subscription time. If the invoice is paid or partially paid, the unused paid amount is refunded; any unpaid unused amount is credited back to the customer.
                                            - 'skip': No credit note is generated for the unused subscription time.

                                            _Note: This field is only applicable to pay-in-advance plans and is ignored for pay-in-arrears plans._; one of: credit, refund, skip
      --on-termination-invoice string       When a subscription is terminated before the end of its billing period, we generate an invoice for the unbilled usage.
                                            This field allows you to control the behavior of the invoice generation:

                                            - 'generate': An invoice is generated for the unbilled usage.
                                            - 'skip': No invoice is generated for the unbilled usage.; one of: generate, skip
      --status string                       Selects which subscription to terminate for the given 'external_id', based on its status. If the field is not defined, Lago targets only 'active' subscriptions.

                                            - 'pending': cancel a subscription scheduled for a future activation (include 'status=pending').
                                            - 'incomplete': cancel a payment-gated subscription that is still awaiting its activation payment (include 'status=incomplete'). The subscription is canceled with 'cancellation_reason: manual', its invoice is closed and nothing is billed, and any pending payment is canceled on a best-effort basis with the payment provider (some provider statuses cannot be canceled). Applied coupons, credit notes and wallet credits are recredited. Termination behaviours ('on_termination_credit_note', 'on_termination_invoice') are ignored, as an incomplete subscription has never been billed.
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

* [lago subscriptions](lago_subscriptions)	 - Manage Lago subscriptions
