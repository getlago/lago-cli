## lago v2 apply

Reconcile the product catalog with a folder of YAML files

### Synopsis

Read every .yml and .yaml file under DIR and make the organization's product catalog match it: billable metrics, product categories, products, rate cards with their rates, and plans with their applied rate cards. Objects are matched by code. apply prints the plan first, then creates and updates in dependency order.

Pass --dry-run to print the plan without changing anything.

Nothing is deleted unless you pass --prune. apply records the objects it manages in a state file, DIR/.lago-state/<organization id>.json by default; commit it with the YAML. --prune deletes the managed objects the files no longer declare, and never touches an object the folder did not create or adopt. The applied rate cards of a declared plan and the pending rates of a declared rate card belong to that object, so --prune also removes the ones the files dropped. Active and terminated rates are history and are kept. Deleting asks for confirmation; pass --confirm prune in scripts.

```
lago v2 apply DIR [flags]
```

### Examples

```
  lago v2 apply billing/ --dry-run
  lago v2 apply billing/
  lago v2 apply billing/ --prune --confirm prune
```

### Options

```
  -h, --help           help for apply
      --prune          Delete managed objects the files no longer declare
      --state string   State file path (default DIR/.lago-state/<organization id>.json)
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

* [lago v2](lago_v2)	 - Lago API v2 (beta): product catalog, contracts and v2 resources
