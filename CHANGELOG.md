# Changelog

All notable changes are generated from conventional commits at release time. This project follows [Semantic Versioning](https://semver.org/).

## Unreleased

- Bootstrap the generated Lago CLI, secure profiles, resilient transport, raw API access, output formats, and golden billing commands.
- **Breaking (pre-1.0):** `create` and `update` commands now print a terse identifier block (`lago_id`, `external_id`, `code`, `name`) in default table output. `--output json` and `--output yaml` are unchanged and still return the complete resource.
