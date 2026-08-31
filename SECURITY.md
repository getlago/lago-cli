# Security Policy

## Reporting a vulnerability

Do not open a public issue. Use GitHub's private vulnerability reporting for `getlago/lago-cli`, or email [security@getlago.com](mailto:security@getlago.com) with the affected version, reproduction, and impact. Lago will acknowledge a complete report within three business days and coordinate disclosure after a fix is available.

Do not include real Lago API keys or customer data in the report. If a credential may have leaked, revoke it immediately; rewriting Git history is not remediation because forks and caches may retain it.

## Supported versions

After GA, the latest minor release of the current major version receives security fixes. Critical fixes may also be backported to the previous major version during its published support window.

## Release security

Release archives are produced by a trusted GitHub Actions workflow, checksummed, SBOM-attested, and signed with keyless cosign identity. Verify the checksum and signature before installation in a sensitive environment.
