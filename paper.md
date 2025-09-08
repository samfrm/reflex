---
title: "Reflex: Local HTTPS referrer emulation for reproducible web analytics and experimentation"
tags:
  - Go
  - CLI
  - Web
  - Measurement
  - Analytics
  - Referrer
authors:
  - name: "Sam (Abbas) Farahmand Pashaki"
    orcid: "0009-0006-7721-2644"
    affiliation: 1
affiliations:
  - name: "Verlag Der Tagesspiegel"
    index: 1
date: 2025-09-08
bibliography: paper.bib
---

# Summary

Browser referrer behavior is central to analytics, experimentation, and attribution pipelines, but it is difficult to reproduce locally. Browsers often suppress `Referer`/`document.referrer` in common local setups (non‑HTTPS origins, external 302 navigations, or OS/browser trust gaps), which prevents meaningful pre‑deployment validation.

Reflex is a small cross‑platform command‑line tool that emulates a realistic referrer locally. It temporarily maps a chosen referrer host to the developer’s machine, serves a trusted HTTPS page with a configurable redirect method, and then navigates to the target application so the browser sets a real `Referer`. Reflex respects the web platform’s referrer policy semantics [@w3c-referrer-policy] and standard HTTP redirection behavior [@rfc9110], and leverages locally trusted certificates via mkcert [@mkcert].

# Statement of need

Researchers and engineers working on measurement, analytics, and experimentation frequently need to validate that instrumentation and targeting work under specific referrer conditions (e.g., partner/campaign origins, paywall previews). In practice, local environments rarely produce the same referrer behavior as production, leading to brittle tests and late discovery of issues (e.g., "Direct/None" sessions, targeting rules that never fire).

Reflex closes this gap by providing a minimal, reproducible workflow to preview an application under a specific referrer, with explicit control over redirect method and `Referrer-Policy`. This enables:

- Pre‑deployment validation of analytics/attribution flows with realistic referrers.
- Reproducible QA of referrer‑sensitive behavior for A/B targeting and paywall gating.
- Repeatable experiments comparing `meta`, `js`, and `302` redirects across browsers.

The tool is intentionally focused: it does not alter browser settings, require extensions, or act as a full proxy. Instead, it uses a short, standards‑compliant HTTPS redirect from a local origin that mimics the selected referrer.

# Functionality

Reflex provides a single binary with three subcommands:

- `reflex run`: Set up the hosts mapping, generate a local TLS certificate, start a minimal HTTPS server, and open a browser (optionally in private/incognito mode) to perform the redirect.
- `reflex status`: Report whether a reflex‑managed hosts entry, certificate pair, and lock file are present for a given referrer.
- `reflex cleanup`: Remove reflex‑managed hosts entries and temporary certificates safely.

Key implementation details:

- TLS certificates are generated via mkcert to achieve local trust without custom configuration [@mkcert]. On Linux, Reflex supports a shared `CAROOT` so elevated and non‑elevated processes share the same local CA store.
- Redirect methods include HTML `meta` refresh (default), JavaScript `window.location`, and HTTP `302 Found`, enabling cross‑method comparison of referrer behavior [@rfc9110]. The served page can set `Referrer-Policy` headers according to the W3C specification [@w3c-referrer-policy].
- Hosts file edits are tagged (`# reflex-managed`) and backed up for safe cleanup.
- A simple file lock avoids concurrent runs from clobbering hosts entries.
- Cross‑platform browser launchers drop elevated privileges where appropriate so the browser opens in the user’s desktop session (Linux/macOS), with optional private/incognito flags.

# Quality control

The repository includes unit tests for the HTTPS redirector (verifying method‑specific behavior and `Referrer-Policy` headers), hosts file management (add/remove, idempotency, bulk cleanup), local locking, and mkcert detection. Continuous Integration builds and runs tests on Linux via GitHub Actions.

# State of the field

Developers commonly rely on ad‑hoc local servers, browser extensions, or general‑purpose proxies to test web behavior. While these tools are powerful, they do not specifically target reproducible referrer emulation with standards‑compliant HTTPS and `Referrer-Policy` control. Reflex fills this narrow but frequently encountered gap with a single‑purpose, scriptable CLI tailored to referrer‑dependent analytics and experimentation.

# Availability

Reflex is implemented in Go and distributed as prebuilt binaries for Linux, macOS, and Windows. Source code and issue tracking are available at: https://github.com/samfrm/reflex. An archival release is available on Zenodo (DOI: https://doi.org/10.5281/zenodo.17079395).

# Acknowledgements

We thank the maintainers of mkcert and the W3C Web Application Security Working Group for the specifications and tooling that make safe local HTTPS possible.

# References
