## reflex — HTTPS referrer emulation for analytics testing

**Reflex** is a lightweight CLI tool helps growth, marketing, and analytics teams reliably emulate inbound referrals (e.g., from `https://news.google.com`) in modern browsers. It does this by:

- Mapping a chosen referrer hostname to your machine via the system hosts file
- Serving a locally trusted HTTPS site for that hostname (certs via mkcert)
- Issuing a redirect (302/meta/JS) to your target URL so the browser sends the correct `Referer`

This enables safe, repeatable end-to-end tests of attribution, A/B splits, and analytics pipelines without fragile header spoofing extensions.

Important: This tool modifies your hosts file and runs a local HTTPS server. It restores cleanly on exit, and you can run `reflex cleanup` at any time.

How it works

- Host spoof: Adds a tagged line like `127.0.0.1 news.google.com # reflex-managed` to your hosts file
- HTTPS: Generates a locally trusted certificate for the hostname with `mkcert`
- Redirect: Serves a small HTTPS site on that host which redirects to your target; the browser sends a real `Referer` header per current referrer policies

Install

- Requires Go 1.21+ for building
- Requires `mkcert` (https://github.com/FiloSottile/mkcert)
- Requires sudo/admin privileges (the CLI runs with sudo)

Build from source

- `go build ./cmd/reflex`
- Or: `go run ./cmd/reflex --help`

One-time Setup (Linux)

To avoid the common "no Firefox/Chrome security databases found" issue when running under sudo, pin a shared CAROOT used by both root and your user.

1. Install the local CA into the system trust store and your user’s browser trust store:

- `sudo mkcert -install` # system trust store
- `mkcert -install` # your user’s Firefox/Chromium trust

2. Pin a shared CAROOT so root and user use the same CA:

- `sudo mkdir -p /etc/mkcert`
- `sudo cp -a "$(mkcert -CAROOT)/." /etc/mkcert/`
- `sudo chmod 755 /etc/mkcert && sudo chmod 644 /etc/mkcert/*`

After this, reflex (which runs with sudo) uses `/etc/mkcert` automatically to generate certs that match your user’s trusted CA.

Notes for other platforms:

- macOS: `brew install mkcert nss && sudo mkcert -install` (no pinned CAROOT needed)
- Windows: `choco install mkcert` then `mkcert -install` in an elevated shell

Run with sudo

- reflex requires elevated privileges to modify `/etc/hosts` and bind to port 443.
- Always run commands with sudo, e.g.: `sudo reflex run --referrer ... --target ...`

Quick start

- Spoof Google News and redirect to your site:
  - `reflex run --referrer https://news.google.com --target https://your-app.example`

Commands

- `reflex run`: Start HTTPS server, spoof host, and open browser
- `reflex cleanup`: Remove hosts entry and generated certificates for a host; add `--all` to remove all reflex-managed entries, all certs, and the lock
- `reflex status`: Inspect whether hosts entry and certs exist

Run options

- `--referrer`: Referrer URL or hostname (required)
- `--target`: Target URL to navigate to (required)
- `--method`: Redirect strategy: meta (default), 302, js
- `--delay`: Delay for meta/js methods in ms (default 1500)
- `--port`: TLS port (default 443). Falls back to 8443 if unavailable
- `--no-browser`: Do not open the browser automatically
- `--private`: Open browser in incognito/private mode (default true)
- `--keep-certs`: Keep generated certs after exit
- `--no-hosts`: Do not modify hosts (advanced)
- `--force-unlock`: Forcefully clear a stale lock before starting
- `--referrer-policy`: Explicit Referrer-Policy for the referrer page (default `origin-when-cross-origin`). Use `unsafe-url` to send the full URL to cross-origin targets.

Notes

- Elevated privileges: Binding to 443 and editing the hosts file typically require admin/sudo
- If 443 is busy or you lack privileges, the server falls back to 8443 and opens `https://<host>:8443`
- Modern default referrer policy is `strict-origin-when-cross-origin`, so your target receives at least the origin
- HSTS domains: A valid, trusted certificate is required. `mkcert` provides a locally trusted CA for this purpose. Reflex now runs `mkcert -install` for you (idempotent) before generating certs; if it fails, you’ll see a clear instruction.

Cleanup and safety

- Tagged hosts entries let reflex remove only what it added
- A timestamped backup of the hosts file is created adjacent to the file on first write
- On exit or `reflex cleanup --referrer <host>`, reflex removes the tagged line and generated certs (unless `--keep-certs`)

Troubleshooting

- Browser blocks the page: Ensure `mkcert` is installed and its root CA is trusted
- Empty `document.referrer`: Set `--referrer-policy origin-when-cross-origin` (default) for origin referrers, or `--referrer-policy unsafe-url` to send the full path. You can also try `--method meta` or `--method js` if your stack handles those better.
- Linux note: If you see "no Firefox and/or Chrome/Chromium security databases found" while using sudo, complete the One-time Setup above to pin a shared CAROOT at `/etc/mkcert`.
- Wrong host is opened: Verify the URL shown by the tool and that your browser didn’t cache an older page
- Hosts entry doesn’t apply: Some VPN/enterprise setups override name resolution; try disabling the VPN temporarily

Development

- Code layout:
  - `cmd/reflex`: CLI entrypoint and subcommands
  - `internal/hosts`: Safe hosts file management with tagged lines
  - `internal/certs`: mkcert integration and cert lifecycle
  - `internal/server`: HTTPS server and redirect strategies
  - `internal/browser`: Cross-platform browser opener
  - `internal/util`: Utilities (hostname parsing, port probing, locking)
- Tests: `internal/hosts` and `internal/util` include basic unit tests you can run with `go test ./...`

Roadmap

- Optional DNS-based spoofing mode (no hosts edits)
- CAP_NET_BIND_SERVICE support on Linux to bind 443 without root
- UI control panel and session recorder
- Native installers and signed binaries
