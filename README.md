## reflex — Local HTTPS referrer emulator ⚡️

[![Release](https://img.shields.io/github/v/release/samfrm/reflex?sort=semver)](https://github.com/samfrm/reflex/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/samfrm/reflex/total.svg)](https://github.com/samfrm/reflex/releases)
[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.17081049.svg)](https://doi.org/10.5281/zenodo.17081049)

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/banner/reflex-banner-dark.png">
    <img alt="Reflex — Local HTTPS referrer emulator" src="assets/banner/reflex-banner-light.png" width="100%">
  </picture>
</p>

Emulate real inbound referrers in modern browsers — safely and repeatably.

✨ What it does

- 🧭 Maps a referrer host to your machine (`/etc/hosts`)
- 🔒 Serves a locally trusted HTTPS site (certs via mkcert)
- 🚀 Redirects to your target so the browser sends a real `Referer`

Perfect for attribution tests, analytics pipelines, and E2E growth flows.

⚠️ Requires sudo/admin (modifies `/etc/hosts`, binds 443)

### 🧩 Visual Flow

```text
   🧑‍💻 You click the referrer URL
             │
             ▼
   /etc/hosts ➜ 127.0.0.1   (spoofs referrer host)
             │
             ▼
   🔒 Reflex HTTPS server (mkcert‑trusted)
             │  302 / <meta> / JS → https://your-app.example
             ▼
   🎯 Target site receives Referer: https://news.google.com/
```

### 🚀 Quick Start

1. One‑time — install mkcert local CA (all platforms)

• macOS: `brew install mkcert nss && sudo mkcert -install`

• Windows: `choco install mkcert` then `mkcert -install` (admin PowerShell)

• Linux (also share CA between root and your user so sudo uses the same CA):

```bash
sudo mkcert -install     # install to system trust store
mkcert -install          # install to your user’s Firefox/Chromium trust
sudo mkdir -p /etc/mkcert
sudo cp -a "$(mkcert -CAROOT)/." /etc/mkcert/
sudo chmod 755 /etc/mkcert && sudo chmod 644 /etc/mkcert/*
```

2. Run a referrer → target flow

```bash
sudo reflex run \
  --referrer https://news.google.com \
  --target   https://your-app.example
```

Reflex opens your default browser (as your normal user) in a private window and serves a small referrer page using HTTPS with a trusted local cert.

### 📦 Install / Download

- Download: https://github.com/samfrm/reflex/releases/latest
- Go install (Go 1.21+):

```
go install github.com/samfrm/reflex/cmd/reflex@latest
```

- One‑liner (Linux/macOS):

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -L "https://github.com/samfrm/reflex/releases/latest/download/reflex_${OS}_${ARCH}.tar.gz" -o reflex.tgz
tar xzf reflex.tgz && rm reflex.tgz
sudo mv reflex /usr/local/bin/
reflex --version
```

- Windows: Download the latest `reflex_windows_<arch>.tar.gz` from Releases and place `reflex.exe` on your PATH.

### 🔧 Install / Build

- 🦫 Go 1.21+
- 🔑 `mkcert` in PATH
- 🏗️ Build: `go build ./cmd/reflex`
- 📖 Help: `go run ./cmd/reflex --help`

### 🕹️ Commands

- ▶️ `reflex run` Start HTTPS server, spoof host, open browser
- 🧹 `reflex cleanup` Remove hosts entry and generated certs (add `--all` to wipe everything)
- 🔍 `reflex status` Show current state for a referrer

### 🎛️ Flags you’ll actually use

- 🔗 `--referrer` Referrer URL or host (required)
- 🎯 `--target` Target URL to navigate to (required)
- 🔁 `--method` Redirect: meta (default), 302, js
- 🛡️ `--referrer-policy` `origin-when-cross-origin` (default) or `unsafe-url` for full URL
- 🕶️ `--private` Open browser in incognito/private mode (default true)
- 🚫 `--no-browser` Don’t auto‑open a browser

More:

- ⏱️ `--delay` (meta/js, ms), 🔌 `--port` (default 443, falls back to 8443), 🗂️ `--keep-certs`, 🧪 `--no-hosts`, 🧹 `--force-unlock`

### 🔬 Research examples

- Validate experiment gating locally (referrer → experiment route):

```bash
sudo reflex run \
  --referrer https://news.google.com \
  --target   https://localhost:3000/experiment
```

- Compare redirect methods and observe `document.referrer` differences:

```bash
sudo reflex run --referrer https://news.google.com --target https://localhost:3000 --method meta
sudo reflex run --referrer https://news.google.com --target https://localhost:3000 --method js
sudo reflex run --referrer https://news.google.com --target https://localhost:3000 --method 302
```

- Evaluate `Referrer-Policy` effects (origin vs full URL):

```bash
sudo reflex run --referrer https://news.google.com --target https://localhost:3000 --referrer-policy origin-when-cross-origin
sudo reflex run --referrer https://news.google.com --target https://localhost:3000 --referrer-policy unsafe-url
```

### 🩹 Troubleshooting (fast answers)

- 🥚 Empty `document.referrer`?
  - Use `--method meta` (default) or `--method js`
  - Try `--referrer-policy unsafe-url` for full URL referrers
- 🧪 Linux mkcert warning under sudo (“no Firefox/Chromium DBs”)?
  - Complete the one‑time setup above (shared CAROOT in `/etc/mkcert`)
- 🖥️ Browser didn’t open?
  - Reflex launches the browser as your non‑root user. If DBus/XDG is missing (headless), copy the printed URL and open manually
- 🌐 Hosts entry not taking effect?
  - Check VPNs/enterprise DNS overrides. `sudo reflex status --referrer <host>` helps debug

### 🧼 Safety and cleanup

- 🏷️ Hosts entries are tagged (`# reflex-managed`) for safe removal
- 🕰️ A timestamped hosts backup is written before first modification
- 🧽 `reflex cleanup --referrer <host>` removes the entry and temp certs (unless `--keep-certs`)

### 🧰 Dev notes

Code map:

- 🧩 `cmd/reflex` CLI
- 🗂️ `internal/hosts` Hosts manager
- 🔑 `internal/certs` mkcert bridge (Linux uses `/etc/mkcert`)
- 🔒 `internal/server` HTTPS redirector
- 🌐 `internal/browser` Browser opener (drops sudo → user, incognito)
- 🛠️ `internal/util` Port/lock/helpers

🧪 Tests: `go test ./...` (unit tests generate self‑signed certs; no mkcert required)

### 📚 Citation

If you use Reflex in your work, please cite the archived release:

- Sam (Abbas) Farahmand Pashaki. Reflex — Local HTTPS referrer emulation for reproducible web analytics and experimentation. Zenodo. https://doi.org/10.5281/zenodo.17081049

See also the citation metadata in `CITATION.cff`.

### 🗺️ Roadmap

- 🧭 Optional DNS spoofing mode (no hosts edits)
- 🧷 CAP_NET_BIND_SERVICE on Linux to bind 443 without sudo
- 🖱️ Simple UI control panel / recorder
- 📦 Packages / signed binaries
