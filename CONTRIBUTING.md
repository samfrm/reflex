Contributing to Reflex
======================

Thanks for your interest in contributing! This document describes how to build, test, and propose changes.

Prerequisites
-------------
- Go 1.21+
- mkcert available on your PATH (for running the CLI; unit tests generate self‑signed certs and do not require mkcert)

Getting started
---------------
1) Fork the repository and create a feature branch.

2) Build and test:

```
make build
make test
```

If your environment restricts binding/listening (e.g., CI sandboxes), some network‑related tests may need to be skipped or run locally.

3) Run the CLI locally (requires admin privileges to modify hosts and bind to 443; falls back to 8443 if needed):

```
sudo ./bin/reflex run \
  --referrer https://news.google.com \
  --target   https://your-app.example
```

4) Lint/format:

```
make fmt
```

Submitting changes
------------------
- Keep PRs focused and small. Include a short rationale in the description.
- Add or update tests for any functional changes.
- Update documentation (`README.md` and/or inline help) where behavior or flags change.
- Follow the Code of Conduct (see CODE_OF_CONDUCT.md).

Release process
---------------
- Tags of the form `vX.Y.Z` trigger GoReleaser to build archives for Linux/macOS/Windows.
- Please ensure tests pass on CI prior to release tags.

Security
--------
Please do not file public issues for security‑sensitive reports. Instead, contact the maintainers privately (open an issue requesting a contact if no address is published).

License
-------
By contributing, you agree that your contributions will be licensed under the MIT License (see LICENSE).

