# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in gccli, please report it responsibly. **Do not open a public issue.**

### How to Report

- **Preferred:** Use [GitHub Security Advisories](https://github.com/bpauli/gccli/security/advisories/new) to report the vulnerability privately.
- **Alternative:** Email the maintainer directly via the contact information on their [GitHub profile](https://github.com/bpauli).

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Scope

The following areas are considered in scope for security reports:

- **Credential handling** — Storage and retrieval of Garmin authentication tokens
- **Token storage** — OS keyring interactions (macOS Keychain, Linux Secret Service, file fallback)
- **Authentication flows** — SSO login, OAuth1/OAuth2 token exchange, token refresh
- **Sensitive data exposure** — Unintended logging or leaking of tokens, passwords, or personal data
- **Command injection** — Unsafe handling of user input in CLI commands

Out of scope:

- Garmin Connect API vulnerabilities (report those to Garmin directly)
- Issues requiring physical access to the user's machine
- Social engineering attacks

## Response

- We will acknowledge receipt of your report within **48 hours**.
- We will provide an initial assessment within **7 days**.
- We aim to release a fix for confirmed vulnerabilities within **30 days**, depending on severity and complexity.

## Supported Versions

Security fixes are applied to the latest release. We do not backport fixes to older versions.

| Version | Supported |
| ------- | --------- |
| Latest  | Yes       |
| Older   | No        |
