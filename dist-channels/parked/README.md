# Parked distribution channels

Nothing in this directory is built, published, tested, or referenced by CI. These are
the install channels Lago decided not to support for the CLI's first release, kept here
so re-enabling one is a docs-and-publish change rather than rewriting it from scratch.

The supported channels are exactly three, documented in the README:

- `brew install getlago/tap/lago`
- `curl -fsSL https://getlago.github.io/lago-cli/install.sh | sh`
- `go install github.com/getlago/lago-cli/cmd/lago@latest`

This is a reduction in **channels**, not in platform support. The release still builds
darwin, linux and windows on amd64 and arm64, CI still compiles and smoke-tests that
full matrix, and `go install` works anywhere Go runs, Windows included.

## What is parked and why

| File | Channel | Why it is parked |
| --- | --- | --- |
| `install.ps1` | Windows script install | Same missing-endpoint problem, on a platform with no smoke coverage. |
| `Dockerfile.release` | `ghcr.io/getlago/lago-cli` | A CLI whose whole job is to hold your billing credentials and read your config file is a poor fit for a container. Nobody asked for it. |

Scoop and Winget had no files of their own: they were `scoops:` and `winget:` blocks in
`.goreleaser.yml` and token wiring in `.github/workflows/release.yml`. Recovering them
means reading those blocks out of the commit that removed them.

The shell installer was parked here from 2026-09-01 to 2026-09-16 and is back at
`install.sh` in the repository root, served from GitHub Pages and smoke-tested from that
URL on every release. It was parked for a missing endpoint, not for a defect; see
DECISIONS.md, "Shell installer, hosted on GitHub Pages", for how the criteria below were
met.

## Re-enable criteria

A channel moves back only when all four hold, in this order:

1. Someone asked. A channel with no demonstrated demand is maintenance with no user.
2. The endpoint or repository is live and Lago controls it. For `install.ps1` that means
   a URL under a Lago-controlled domain or the `getlago` GitHub organization returning
   HTTP 200; `get.lago.com` is not Lago's and must never appear in an artifact.
3. A post-release smoke job installs from the real endpoint on every release and fails
   the release when it cannot. A channel with no smoke test is an untested channel.
4. The README documents it only after that job has passed once. Publish, verify,
   document, in that order.

`lago upgrade` does not replace the running binary for any channel, the shell installer
included: it prints the command for how the binary was installed, and for a script
install that command is the installer itself, which is idempotent. A PowerShell channel
would follow the same rule rather than restoring the removed self-replace path.

See DECISIONS.md, "Two install channels for 1.0".
