# Parked distribution channels

Nothing in this directory is built, published, tested, or referenced by CI. These are
the install channels Lago decided not to support for the CLI's first release, kept here
so re-enabling one is a docs-and-publish change rather than rewriting it from scratch.

The supported channels are exactly two, documented in the README:

- `brew install getlago/tap/lago`
- `go install github.com/getlago/lago-cli/cmd/lago@latest`

This is a reduction in **channels**, not in platform support. The release still builds
darwin, linux and windows on amd64 and arm64, CI still compiles and smoke-tests that
full matrix, and `go install` works anywhere Go runs, Windows included.

## What is parked and why

| File | Channel | Why it is parked |
| --- | --- | --- |
| `install.sh` | `curl -fsSL https://getlago.com/install.sh \| sh` | The endpoint was never published. A `curl \| sh` against a missing URL installs nothing and exits 0, so the channel was a silent failure waiting for its first user. |
| `install.ps1` | Windows script install | Same missing-endpoint problem, on a platform with no smoke coverage. |
| `Dockerfile.release` | `ghcr.io/getlago/lago-cli` | A CLI whose whole job is to hold your billing credentials and read your config file is a poor fit for a container. Nobody asked for it. |

Scoop and Winget had no files of their own: they were `scoops:` and `winget:` blocks in
`.goreleaser.yml` and token wiring in `.github/workflows/release.yml`. Recovering them
means reading those blocks out of the commit that removed them.

## Re-enable criteria

A channel moves back only when all four hold, in this order:

1. Someone asked. A channel with no demonstrated demand is maintenance with no user.
2. The endpoint or repository is live and Lago controls it. For `install.sh` that means
   `https://getlago.com/install.sh` returning HTTP 200 from a Lago-controlled domain;
   `get.lago.com` is not Lago's and must never appear in an artifact.
3. A post-release smoke job installs from the real endpoint on every release and fails
   the release when it cannot. A channel with no smoke test is an untested channel.
4. The README documents it only after that job has passed once. Publish, verify,
   document, in that order.

Self-update is a fifth condition for any script channel. `lago upgrade` no longer
replaces the running binary: with only Homebrew and `go install` supported, there is no
install that the CLI itself owns, so `upgrade` prints the command for how the binary was
installed. Restoring a script channel means restoring the download, checksum-verify and
atomic-replace path that was removed with it, and the signature verification that path
depended on.

See DECISIONS.md, "Two install channels for 1.0".
