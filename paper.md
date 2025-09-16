---
title: "Reflex: Local HTTPS referrer emulation for reproducible web analytics and experimentation"
tags:
  - Go
  - CLI
  - Web
  - Measurement
  - Analytics
  - Referrer
  - Reproducibility
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

Browser referrer behavior is central to analytics, experimentation, and attribution pipelines, but it is difficult to reproduce locally in a way that is faithful to production and repeatable by other researchers. Browsers often suppress `Referer`/`document.referrer` in common local setups (non‑HTTPS origins, external 302 navigations, or OS/browser trust gaps), which prevents meaningful pre‑deployment validation and undermines the reproducibility of measurement.

Reflex is a small cross‑platform command‑line tool that emulates a realistic referrer locally. It temporarily maps a chosen referrer host to the developer’s machine, serves a trusted HTTPS page with a configurable redirect method, and then navigates to the target application so the browser sets a real `Referer`. Reflex respects the web platform’s referrer policy semantics [@w3c-referrer-policy] and standard HTTP redirection behavior [@rfc9110], and leverages locally trusted certificates via mkcert [@mkcert]. This enables researchers and engineers to reproduce and share referrer‑dependent scenarios in a single command, improving the validity of analytics and experiment instrumentation.

# Statement of need

Researchers and engineers working on measurement, analytics, and experimentation frequently need to validate that instrumentation and targeting work under specific referrer conditions (e.g., partner/campaign origins, paywall previews). In practice, local environments rarely produce the same referrer behavior as production, leading to brittle tests and late discovery of issues (e.g., "Direct/None" sessions, targeting rules that never fire).

Reflex closes this gap by providing a minimal, reproducible workflow to preview an application under a specific referrer, with explicit control over redirect method and `Referrer-Policy`. This enables:

- Pre‑deployment validation of analytics/attribution flows with realistic referrers.
- Reproducible QA of referrer‑sensitive behavior for A/B targeting and paywall gating.
- Repeatable experiments comparing `meta`, `js`, and `302` redirects across browsers.

The tool is intentionally focused: it does not alter browser settings, require extensions, or act as a full proxy. Instead, it uses a short, standards‑compliant HTTPS redirect from a local origin that mimics the selected referrer.

Minimal reproducible example. The following single command reproduces a realistic referrer across OSes using locally trusted HTTPS:

```
reflex run --referrer https://news.google.com --target https://localhost:3000/experiment
```

This opens a private/incognito browser window and performs a standards‑conformant redirect so the target observes a populated Referer (e.g., `https://news.google.com/`). The same invocation can be shared to reproduce the scenario on other machines.



# Use cases for research

- Analytics/attribution validation for controlled experiments: Researchers running online controlled experiments (A/B tests) often need to verify instrumentation and gating rules that depend on referrer context (e.g., arriving from a partner site). Reflex provides a reproducible way to preview and test these flows locally before running or analyzing experiments [@kohavi2020trustworthy].
- Cross‑browser measurement of `Referrer-Policy` and redirect methods: Web measurement studies can compare how browsers populate `Referer`/`document.referrer` under `meta`, `js`, and `302` redirects with different policies. Reflex enables these controlled tests across OSes with locally trusted TLS [@w3c-referrer-policy; @rfc9110].
- Privacy and tracking studies: Prior work shows the prevalence and risks of referrer‑based tracking and leakage on the web [@englehardt2016online; @acar2014never]. Reflex can generate reproducible referrer scenarios to validate mitigations (e.g., policy choices) or to build small‑scale datasets illustrating leakage conditions.

Expected impact. While seemingly narrow, reliable referrer emulation is a recurring prerequisite for trustworthy analytics, A/B testing, and privacy measurement. Reflex lowers the setup cost for these workflows and makes them repeatable across environments, which can reduce instrumentation regressions and improve the integrity of subsequent analyses [@kohavi2020trustworthy].

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

# Limitations and scope

Reflex is intentionally narrow in scope:

- Not a proxy or MITM: it does not intercept traffic or modify requests beyond serving a short redirect.
- Local emulation only: it is designed for local testing, not production spoofing of referrers.
- No browser reconfiguration: it avoids changing browser settings, profiles, or policies.
- Subject to site controls: site‑specific defenses (e.g., CSP, framebusting) remain in effect.

# Quality control

The repository includes unit tests for the HTTPS redirector (verifying method‑specific behavior and `Referrer-Policy` headers), hosts file management (add/remove, idempotency, bulk cleanup), local locking, and mkcert detection. Tests generate ephemeral self‑signed certificates (no mkcert required) to isolate server behavior. Continuous Integration builds and runs tests on Linux, macOS, and Windows via GitHub Actions.

# State of the field

Developers commonly rely on ad‑hoc local servers, browser extensions, or general‑purpose proxies (e.g., mitmproxy [@mitmproxy]) to test web behavior. Automation frameworks such as Selenium [@selenium] and Playwright [@playwright] can orchestrate browser flows, but they do not specifically target reproducible referrer emulation with locally trusted HTTPS and precise `Referrer-Policy` control out‑of‑the‑box. Reflex fills this narrow but frequently encountered gap with a single‑purpose, scriptable CLI tailored to referrer‑dependent analytics and experimentation. Unlike browser extensions and flag‑based overrides, it produces a standards‑conformant HTTPS navigation with a trusted certificate and explicit `Referrer‑Policy`, yielding consistent, shareable results across browsers and OSes.

# Availability

Reflex is implemented in Go and distributed as prebuilt binaries for Linux, macOS, and Windows. Source code and issue tracking are available at: https://github.com/samfrm/reflex. Reflex is MIT‑licensed (LICENSE: https://github.com/samfrm/reflex/blob/main/LICENSE) and archived per‑version on Zenodo (DOI: https://doi.org/10.5281/zenodo.17081049). This paper refers to version v0.1.3 of the software.

# History and maturity

Reflex packages a pattern I’ve used for years—local HTTPS, a short redirect, and hosts mapping—into a single CLI. I open‑sourced and packaged it recently, which is why the public commit history is short. Compared to the original scripts, the CLI adds: mkcert‑based local trust with a shared CAROOT on Linux (/etc/mkcert), tagged and backed‑up /etc/hosts edits (“# reflex‑managed”), a file lock to prevent concurrent runs, and a browser launcher that drops privileges to open a private window in the user session. The repository includes tests and CI; server tests assert the `Referrer‑Policy` and redirect bodies/headers for `meta`, `js`, and `302` methods. I’m not aware of citations yet due to the recent release, but Reflex directly supports the reproducible analytics/experimentation and privacy measurement workflows described in this paper.

# Acknowledgements

We thank the maintainers of mkcert and the W3C Web Application Security Working Group for the specifications and tooling that make safe local HTTPS possible.

# References
