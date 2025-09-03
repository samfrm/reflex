# 🪞 Reflex

**Simulate Referrer-Based Traffic with HTTPS Redirects**

**Reflex** is a lightweight CLI tool for developers and QA engineers to simulate web traffic that appears to originate from trusted domains like `news.google.com`. It’s ideal for testing A/B experiments, analytics triggers (e.g. Piano, Adobe Target), or any system that relies on referrer headers.

> ⚡️ _“Redirect. Spoof. Test. Repeat.”_

---

## 🚀 Features

- ✅ Simulates traffic with any `Referer` value
- 🔒 HTTPS redirect server with trusted local certs (via `mkcert`)
- 🛠️ Safe modification & restoration of `/etc/hosts`
- 🌐 Automatically opens browser to spoofed domain
- 🔁 Auto-fallback from port 443 → 8443
- 🧹 Cleanup certs and hosts on exit
- 🐧 macOS + Linux support

---

## 🔧 Use Cases

- ✅ Test A/B testing platforms (Adobe Target, Optimizely, etc.)
- ✅ Validate analytics pipelines (Piano, GA, Segment, etc.)
- ✅ Simulate traffic from social media, news aggregators, or search engines
- ✅ Reproduce edge-case behaviors (like `document.referrer` logic)

---
