# Changelog

All notable changes are generated from conventional commits at release time. This project follows [Semantic Versioning](https://semver.org/).

## Unreleased

- Bootstrap the generated Lago CLI, secure profiles, resilient transport, raw API access, output formats, and golden billing commands.
- **Breaking (pre-1.0):** install channels are Homebrew (`brew install getlago/tap/lago`) and `go install` only. The shell and PowerShell installers, the GHCR image, Scoop, and Winget are parked in `dist-channels/parked/`. Platform support is unchanged: all six os/arch targets are still built and smoke-tested.
- **Breaking (pre-1.0):** `lago upgrade` prints the upgrade command for how the binary was installed instead of replacing it in place.
- **Breaking (pre-1.0):** `create` and `update` commands now print a terse identifier block (`lago_id`, `external_id`, `code`, `name`) in default table output. `--output json` and `--output yaml` are unchanged and still return the complete resource.
