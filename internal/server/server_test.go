package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// genSelfSigned writes a localhost cert/key pair to dir and returns their paths.
func genSelfSigned(t *testing.T, dir string) (string, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	return certFile, keyFile
}

func startTLS(t *testing.T, cfg Config) (addr string, shutdown func()) {
	t.Helper()
	srv, err := NewHTTPServer(cfg)
	if err != nil {
		t.Fatalf("NewHTTPServer: %v", err)
	}
	// Bind ephemeral port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen :0: %v", err)
	}
	tcp := ln.Addr().(*net.TCPAddr)
	addr = fmt.Sprintf("127.0.0.1:%d", tcp.Port)
	go func() { _ = srv.ServeTLS(ln, cfg.CertFile, cfg.KeyFile) }()
	shutdown = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
	return addr, shutdown
}

func httpClientInsecure() *http.Client {
    tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
    c := &http.Client{Transport: tr, Timeout: 3 * time.Second}
    // Do not follow redirects so tests can assert the initial 302 response
    c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    }
    return c
}

func TestServer302(t *testing.T) {
	dir := t.TempDir()
	cert, key := genSelfSigned(t, dir)
	cfg := Config{Port: 0, CertFile: cert, KeyFile: key, Method: Method302, Target: "https://example.com/target", ReferrerPolicy: "origin-when-cross-origin"}
	addr, stop := startTLS(t, cfg)
	defer stop()

	c := httpClientInsecure()
	req, _ := http.NewRequest("GET", "https://"+addr+"/", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("do 302: %v", err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status=%d want 302", resp.StatusCode)
	}
	if got := resp.Header.Get("Referrer-Policy"); got != "origin-when-cross-origin" {
		t.Fatalf("Referrer-Policy=%q", got)
	}
	if got := resp.Header.Get("Location"); got != cfg.Target {
		t.Fatalf("Location=%q want %q", got, cfg.Target)
	}
}

func TestServerMetaAndJS(t *testing.T) {
	dir := t.TempDir()
	cert, key := genSelfSigned(t, dir)
	for _, m := range []RedirectMethod{MethodMeta, MethodJS} {
		cfg := Config{Port: 0, CertFile: cert, KeyFile: key, Method: m, Target: "https://example.com/target", Delay: 500 * time.Millisecond, ReferrerPolicy: "unsafe-url"}
		addr, stop := startTLS(t, cfg)
		c := httpClientInsecure()
		resp, err := c.Get("https://" + addr + "/")
		if err != nil {
			stop()
			t.Fatalf("get %s: %v", m, err)
		}
		if resp.StatusCode != 200 {
			stop()
			t.Fatalf("status=%d want 200", resp.StatusCode)
		}
		if got := resp.Header.Get("Referrer-Policy"); got != "unsafe-url" {
			stop()
			t.Fatalf("Referrer-Policy=%q", got)
		}
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		s := string(b)
		if !strings.Contains(s, cfg.Target) {
			stop()
			t.Fatalf("body missing target URL for %s", m)
		}
		if !strings.Contains(s, "meta name=\"referrer\"") {
			stop()
			t.Fatalf("body missing referrer meta for %s", m)
		}
		stop()
	}
}
