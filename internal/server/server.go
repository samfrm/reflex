package server

import (
    "fmt"
    "log"
    "net/http"
    "time"
)

type RedirectMethod string

const (
	Method302  RedirectMethod = "302"
	MethodMeta RedirectMethod = "meta"
	MethodJS   RedirectMethod = "js"
)

type Config struct {
    Port       int
    CertFile   string
    KeyFile    string
    Method     RedirectMethod
    Target     string
    Delay      time.Duration
    RefHost    string
    LogVerbose bool
    ReferrerPolicy string
}

// NewHTTPServer builds an *http.Server with a dedicated handler for the
// provided configuration. Tests can use this to start/stop the server.
func NewHTTPServer(cfg Config) (*http.Server, error) {
    mux := http.NewServeMux()
    switch cfg.Method {
    case Method302:
        mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            if cfg.ReferrerPolicy != "" {
                w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
            }
            http.Redirect(w, r, cfg.Target, http.StatusFound)
        })
    case MethodMeta:
        mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            if cfg.ReferrerPolicy != "" {
                w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
            }
            fmt.Fprintf(w, `<!doctype html><html><head><title>Redirect</title><meta name="referrer" content="%s"><meta http-equiv="refresh" content="%.1f;url=%s"></head><body>Redirecting to <a href="%s">target</a>…</body></html>`, cfg.ReferrerPolicy, cfg.Delay.Seconds(), cfg.Target, cfg.Target)
        })
    case MethodJS:
        mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            if cfg.ReferrerPolicy != "" {
                w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
            }
            fmt.Fprintf(w, `<!doctype html><html><head><title>Redirect</title><meta name="referrer" content="%s"></head><body>Redirecting to <a id="l" href="%s">target</a>…<script>setTimeout(function(){window.location=%q}, %d)</script></body></html>`, cfg.ReferrerPolicy, cfg.Target, cfg.Target, int(cfg.Delay.Milliseconds()))
        })
    default:
        return nil, fmt.Errorf("unknown redirect method: %s", cfg.Method)
    }
    addr := fmt.Sprintf(":%d", cfg.Port)
    return &http.Server{Addr: addr, Handler: mux}, nil
}

func Run(cfg Config) error {
    srv, err := NewHTTPServer(cfg)
    if err != nil {
        return err
    }
    addr := fmt.Sprintf(":%d", cfg.Port)
    log.Printf("starting HTTPS server on %s", addr)
    return srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
}
