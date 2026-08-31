# Contributing

Thank you for improving Lago CLI. This is a public-source project even while repository visibility is being prepared; never put a real key, customer, invoice, email, internal hostname, staging URL, or unreleased roadmap detail in a commit, fixture, issue, or CI log.

1. Branch from `main` and use conventional commits.
2. Update `spec/openapi.yaml` only from the public Lago OpenAPI source. Run `make generate` and commit every generated result.
3. Add a regression test for every behavior change.
4. Run `make test coverage lint security generate-check`.
5. Open a pull request using the template. Maintainers review generated command and JSON-contract changes as breaking-surface changes.

Use obviously fake credentials such as `lago_test_FAKE000000000000000000000000`. Nightly staging credentials exist only in the protected GitHub environment and are unavailable to fork pull requests.

API ambiguity belongs in the upstream OpenAPI repository. Do not guess billing semantics or add local operation/type mappings.
