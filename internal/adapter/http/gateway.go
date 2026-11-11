// Package http hosts the grpc-gateway reverse proxy: an HTTP/JSON server that
// translates REST calls into gRPC calls against the service's own gRPC server.
// It exists so the service can be exercised with plain HTTP tools (curl, Postman)
// in addition to gRPC clients.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	allnetv1 "github.com/Duke40199/sega-idac-client/gen/allnetv1"
	healthv1 "github.com/Duke40199/sega-idac-client/gen/healthv1"
	segaidacv1 "github.com/Duke40199/sega-idac-client/gen/segaidacv1"
	"github.com/Duke40199/sega-idac-client/internal/config"
)

// handshakeNoiseWriter drops pure connection-abort TLS handshake log lines —
// clients that open a socket and close it without completing the handshake
// (browser preconnect sockets, TCP health/port probes, dropped connections).
// These are logged by net/http as "TLS handshake error ... EOF" (or connection
// reset / i/o timeout) once per aborted connection and are not server faults.
// Actionable handshake errors (untrusted cert, HTTP-sent-to-HTTPS) are kept.
type handshakeNoiseWriter struct{ out io.Writer }

func (w handshakeNoiseWriter) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte("TLS handshake error")) {
		for _, noise := range [][]byte{[]byte("EOF"), []byte("connection reset by peer"), []byte("i/o timeout")} {
			if bytes.Contains(p, noise) {
				return len(p), nil // swallow benign connection-abort noise
			}
		}
	}
	return w.out.Write(p)
}

// routingErrorHandler writes a clean JSON error for unmatched routes / method
// mismatches (404, 501, 400). The grpc-gateway default routes these through
// HTTPError with a context that never carries ServerMetadata, which logs a
// misleading "Failed to extract ServerMetadata from context" error on every
// 404 (e.g. a browser's /favicon.ico or a trailing-slash path). Handling them
// here keeps the correct status while dropping that noise.
func routingErrorHandler(_ context.Context, _ *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, _ *http.Request, httpStatus int) {
	w.Header().Set("Content-Type", m.ContentType(nil))
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    httpStatus,
		"message": http.StatusText(httpStatus),
	})
}

// Gateway wraps the HTTP reverse-proxy server.
type Gateway struct {
	server  *http.Server
	addr    string
	tlsCert string
	tlsKey  string
	tlsOn   bool
}

// NewGateway builds an HTTP/JSON gateway that forwards to the local gRPC server.
// The provided ctx governs the lifetime of the upstream gRPC connections.
func NewGateway(ctx context.Context, cfg config.Config) (*Gateway, error) {
	mux := runtime.NewServeMux(
		runtime.WithRoutingErrorHandler(routingErrorHandler),
	)
	endpoint := cfg.GetGRPCCfg().DialAddr()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := healthv1.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		return nil, err
	}
	if err := segaidacv1.RegisterSegaIDACServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		return nil, err
	}
	if err := allnetv1.RegisterAllNetServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		return nil, err
	}

	httpCfg := cfg.GetHTTPCfg()
	addr := httpCfg.Addr()
	return &Gateway{
		server: &http.Server{
			Addr:     addr,
			Handler:  mux,
			ErrorLog: log.New(handshakeNoiseWriter{out: os.Stderr}, "", log.LstdFlags),
		},
		addr:    addr,
		tlsCert: httpCfg.TLSCertFile,
		tlsKey:  httpCfg.TLSKeyFile,
		tlsOn:   httpCfg.TLSEnabled(),
	}, nil
}

// Start serves the gateway. It serves HTTPS when a TLS cert/key were configured,
// otherwise plain HTTP. It blocks until the server stops; http.ErrServerClosed is
// returned on a graceful shutdown and should be treated as success by the caller.
func (g *Gateway) Start() error {
	if g.tlsOn {
		log.Printf("🔒 HTTPS gateway listening on %s", g.addr)
		return g.server.ListenAndServeTLS(g.tlsCert, g.tlsKey)
	}
	log.Printf("🌐 HTTP gateway listening on %s", g.addr)
	return g.server.ListenAndServe()
}

// Stop gracefully shuts down the HTTP gateway.
func (g *Gateway) Stop(ctx context.Context) error {
	return g.server.Shutdown(ctx)
}
