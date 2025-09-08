package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/samfrm/reflex/internal/browser"
	"github.com/samfrm/reflex/internal/certs"
	"github.com/samfrm/reflex/internal/hosts"
	"github.com/samfrm/reflex/internal/server"
	"github.com/samfrm/reflex/internal/util"
)

const (
	defaultIP           = "127.0.0.1"
	defaultPortTLS      = 443
	defaultFallbackPort = 8443
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usageAndExit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "run":
		if err := util.RequireRoot(); err != nil {
			// Print a clear warning early and exit without extra "error:" noise.
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if err := runCmd(os.Args[2:]); err != nil {
			log.Printf("error: %v", err)
			os.Exit(1)
		}
	case "cleanup":
		if err := util.RequireRoot(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if err := cleanupCmd(os.Args[2:]); err != nil {
			log.Printf("error: %v", err)
			os.Exit(1)
		}
	case "status":
		if err := util.RequireRoot(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if err := statusCmd(os.Args[2:]); err != nil {
			log.Printf("error: %v", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usageAndExit(0)
	case "version", "-v", "--version":
		fmt.Println("reflex v0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usageAndExit(2)
	}
}

func usageAndExit(code int) {
	fmt.Fprintf(os.Stderr, `
reflex — Local HTTPS referrer emulation for analytics testing

Usage:
  reflex <command> [options]

Commands:
  run       Start HTTPS server, spoof host, open browser
  cleanup   Remove host mapping and generated certs
  status    Show current state for a referrer

Examples:
  reflex run --referrer https://news.google.com --target https://example.com
  reflex cleanup --referrer news.google.com
  reflex status --referrer news.google.com

Use "reflex <command> -h" for command-specific help.
`)
	os.Exit(code)
}

func runCmd(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	referrer := fs.String("referrer", "", "Referrer URL or hostname (e.g., https://news.google.com)")
	target := fs.String("target", "", "Target URL to navigate to")
	ip := fs.String("ip", defaultIP, "IP to map the referrer host to")
	port := fs.Int("port", defaultPortTLS, "TLS port to serve on (443 requires elevated privileges)")
	fallbackPort := fs.Int("fallback-port", defaultFallbackPort, "Fallback port if desired port is unavailable")
	method := fs.String("method", "meta", "Redirect method: meta|302|js (default meta)")
	refPol := fs.String("referrer-policy", "origin-when-cross-origin", "Referrer-Policy to use (e.g., no-referrer, origin, origin-when-cross-origin, strict-origin-when-cross-origin, unsafe-url)")
	delay := fs.Int("delay", 1500, "Delay in ms for meta/js redirect methods")
    noBrowser := fs.Bool("no-browser", false, "Do not open the browser automatically")
    private := fs.Bool("private", true, "Open browser in incognito/private mode")
	keepCerts := fs.Bool("keep-certs", false, "Keep generated certificates after exit")
	noHosts := fs.Bool("no-hosts", false, "Do not modify hosts file (advanced)")
	hostsPath := fs.String("hosts-file", "", "Override hosts file path (testing)")
	certDir := fs.String("cert-dir", "", "Directory to write certs to (defaults to temp)")
	duration := fs.Duration("duration", 0, "Optional auto-shutdown duration (e.g., 5m, 1h)")
	verbose := fs.Bool("verbose", false, "Verbose logs")
	forceUnlock := fs.Bool("force-unlock", false, "Forcefully remove an existing lock before starting")
	_ = fs.Parse(args)

	if *referrer == "" || *target == "" {
		fs.Usage()
		return fmt.Errorf("missing required flags: --referrer and --target")
	}

	if *verbose {
		util.EnableVerbose()
	}

	host, herr := util.ExtractHostname(*referrer)
	if herr != nil {
		return fmt.Errorf("invalid --referrer: %w", herr)
	}

	if !strings.EqualFold(*method, "302") && !strings.EqualFold(*method, "meta") && !strings.EqualFold(*method, "js") {
		return fmt.Errorf("invalid --method: %s", *method)
	}

	// Preflight: mkcert presence. Do not run `mkcert -install` here; that is a one-time setup.
	if !certs.IsMkcertInstalled() {
		log.Println("mkcert not found. Install from https://github.com/FiloSottile/mkcert")
		switch runtime.GOOS {
		case "darwin":
			log.Println("Tip (macOS): brew install mkcert nss && sudo mkcert -install")
		case "linux":
			log.Println("Tip (Linux): Debian/Ubuntu → sudo apt-get install mkcert libnss3-tools; Fedora → sudo dnf install mkcert nss-tools; Arch → sudo pacman -S mkcert nss")
		case "windows":
			log.Println("Tip (Windows): choco install mkcert, then run mkcert -install in an elevated shell")
		}
		return fmt.Errorf("mkcert is required to create a locally trusted cert for HTTPS referrer emulation")
	}
	// On Linux we require a shared CAROOT so root and user share the same CA.
	var pinnedCAROOT string
	if runtime.GOOS == "linux" {
		pinnedCAROOT = "/etc/mkcert"
		if !util.PathExists(pinnedCAROOT) || !util.PathExists(filepath.Join(pinnedCAROOT, "rootCA.pem")) {
			return fmt.Errorf("missing pinned CAROOT at %s. Run the one-time setup:\n sudo mkcert -install\n  mkcert -install\n  sudo mkdir -p /etc/mkcert\n  sudo cp -a \"$(mkcert -CAROOT)/.\" /etc/mkcert/\n  sudo chmod 755 /etc/mkcert && sudo chmod 644 /etc/mkcert/*\nThen re-run this command with sudo", pinnedCAROOT)
		}
	}

	// Lock to prevent concurrent runs from clobbering hosts
	if *forceUnlock {
		_ = util.RemoveLock()
	}
	lock, lerr := util.AcquireLock()
	if lerr != nil {
		return lerr
	}
	// Ensure release on normal returns
	defer lock.Release()

	// Determine cert directory
	dir := *certDir
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "reflex", host)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("create cert dir: %v", err)
	}

	// Setup cleanup signals
	var addedHost bool
	cleanup := func() {
		if addedHost && !*noHosts {
			_ = hosts.Manager{Path: hosts.PathOrDefault(*hostsPath)}.Remove(host)
		}
		if !*keepCerts {
			_ = os.RemoveAll(dir)
		}
		// Always try to release lock (idempotent)
		lock.Release()
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

	// Hosts modification
	if !*noHosts {
		mgr := hosts.Manager{Path: hosts.PathOrDefault(*hostsPath)}
		if err := mgr.Add(*ip, host); err != nil {
			if errors.Is(err, hosts.ErrAlreadyPresent) {
				util.VLog("hosts entry already present")
			} else {
				return fmt.Errorf("update hosts: %w", err)
			}
		} else {
			addedHost = true
		}
	} else {
		util.VLog("--no-hosts enabled; not touching hosts file")
	}

	// Cert generation via mkcert (CA is ensured already)
	var certFile, keyFile string
	var err error
	if pinnedCAROOT != "" {
		certFile, keyFile, err = certs.EnsureCertificatesWithCAROOT(host, dir, pinnedCAROOT)
	} else {
		certFile, keyFile, err = certs.EnsureCertificates(host, dir)
	}
	if err != nil {
		return fmt.Errorf("generate certificates: %w", err)
	}

	// Port selection
	p := *port
	if !util.CanBind(p) {
		log.Printf("port %d unavailable; falling back to %d", p, *fallbackPort)
		p = *fallbackPort
		if !util.CanBind(p) {
			return fmt.Errorf("fallback port %d also unavailable", p)
		}
	}

	// Start server
	srv := server.Config{
		Port:           p,
		CertFile:       certFile,
		KeyFile:        keyFile,
		Method:         server.RedirectMethod(strings.ToLower(*method)),
		Target:         *target,
		Delay:          time.Duration(*delay) * time.Millisecond,
		RefHost:        host,
		LogVerbose:     *verbose,
		ReferrerPolicy: *refPol,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- server.Run(srv) }()

	// Compose URL and open browser
	url := fmt.Sprintf("https://%s", host)
	if p != 443 {
		url = fmt.Sprintf("%s:%d", url, p)
	}
	log.Printf("serving spoofed referrer at %s", url)
	if strings.EqualFold(*method, "302") {
		log.Printf("Heads-up: 302 redirects from an external open may yield empty document.referrer in some browsers. For consistent results, use --method meta or --method js.")
	}
    if !*noBrowser {
        if err := browser.Open(url, *private); err != nil {
            log.Printf("open browser: %v", err)
            log.Printf("Please open this URL manually: %s. (private-mode recommended)", url)
        }
    } else {
        log.Printf("Open this URL in your browser: %s. (private-mode recommended)", url)
    }

	if *duration > 0 {
		log.Printf("auto-shutdown after %s", *duration)
		select {
		case <-time.After(*duration):
			cleanup()
			return nil
		case err := <-errCh:
			cleanup()
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Block until server exits or signal triggers cleanup
	if err := <-errCh; err != nil {
		cleanup()
		return fmt.Errorf("server error: %w", err)
	}
	cleanup()
	return nil
}

func cleanupCmd(args []string) error {
	fs := flag.NewFlagSet("cleanup", flag.ExitOnError)
	referrer := fs.String("referrer", "", "Referrer host or URL whose mapping to remove")
	hostsPath := fs.String("hosts-file", "", "Override hosts file path (testing)")
	certDir := fs.String("cert-dir", "", "Certificate directory to remove (default temp per referrer)")
	keepCerts := fs.Bool("keep-certs", false, "Keep certificates; only remove hosts mapping")
	all := fs.Bool("all", false, "Remove all reflex-managed hosts entries, all temp certs, and the lock")
	_ = fs.Parse(args)

	if *referrer == "" && !*all {
		fs.Usage()
		return fmt.Errorf("specify --referrer or --all")
	}
	mgr := hosts.Manager{Path: hosts.PathOrDefault(*hostsPath)}

	if *all {
		n, err := mgr.RemoveAllTagged()
		if err != nil {
			log.Printf("remove hosts entries: %v", err)
		} else {
			log.Printf("removed %d hosts entrie(s)", n)
		}
		if !*keepCerts {
			base := filepath.Join(os.TempDir(), "reflex")
			if err := os.RemoveAll(base); err != nil {
				log.Printf("remove certs base: %v", err)
			} else {
				log.Printf("removed certs base at %s", base)
			}
		}
		if err := util.RemoveLock(); err == nil {
			log.Printf("removed lock file")
		}
		return nil
	}

	host, err := util.ExtractHostname(*referrer)
	if err != nil {
		return fmt.Errorf("invalid --referrer: %w", err)
	}
	if err := mgr.Remove(host); err != nil {
		log.Printf("remove hosts entry: %v", err)
	} else {
		log.Printf("removed hosts entry for %s", host)
	}

	if !*keepCerts {
		dir := *certDir
		if dir == "" {
			dir = filepath.Join(os.TempDir(), "reflex", host)
		}
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("remove certs: %v", err)
		} else {
			log.Printf("removed certs at %s", dir)
		}
	}
	if err := util.RemoveLock(); err == nil {
		log.Printf("removed lock file")
	}
	return nil
}

func statusCmd(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	referrer := fs.String("referrer", "", "Referrer host or URL to check")
	hostsPath := fs.String("hosts-file", "", "Override hosts file path (testing)")
	_ = fs.Parse(args)

	if *referrer == "" {
		fs.Usage()
		return fmt.Errorf("missing --referrer")
	}
	host, err := util.ExtractHostname(*referrer)
	if err != nil {
		return fmt.Errorf("invalid --referrer: %w", err)
	}

	mgr := hosts.Manager{Path: hosts.PathOrDefault(*hostsPath)}
	present, line, err := mgr.Contains(host)
	if err != nil {
		return fmt.Errorf("read hosts: %w", err)
	}
	if present {
		fmt.Printf("hosts entry present: %s\n", strings.TrimSpace(line))
	} else {
		fmt.Println("hosts entry not present")
	}

	dir := filepath.Join(os.TempDir(), "reflex", host)
	if util.PathExists(filepath.Join(dir, "cert.pem")) && util.PathExists(filepath.Join(dir, "key.pem")) {
		fmt.Printf("certs present: %s\n", dir)
	} else {
		fmt.Println("certs not present")
	}
	if util.PathExists(filepath.Join(os.TempDir(), "reflex.lock")) {
		fmt.Println("lock: present")
	} else {
		fmt.Println("lock: not present")
	}
	return nil
}
