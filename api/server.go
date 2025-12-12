package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// Server routes requests to capability handlers.
type Server struct {
	corsOrigin  string
	bearerToken string
	tlsCertFile string
	tlsKeyFile  string
	serve       func(*http.Server) error                 // optional override for tests
	serveTLS    func(*http.Server, string, string) error // optional override for tests
	incident    IncidentHandler
	alert       AlertHandler
	log         LogHandler
	metric      MetricHandler
	ticket      TicketHandler
	messaging   MessagingHandler
	service     ServiceHandler
	deployment  DeploymentHandler
	secret      SecretProvider
}

// NewServerFromEnv constructs a Server with providers loaded from environment variables.
func NewServerFromEnv(ctx context.Context) (*Server, error) {
	corsOrigin := os.Getenv("OPSORCH_CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}
	bearer := strings.TrimSpace(os.Getenv("OPSORCH_BEARER_TOKEN"))
	tlsCertFile := strings.TrimSpace(os.Getenv("OPSORCH_TLS_CERT_FILE"))
	tlsKeyFile := strings.TrimSpace(os.Getenv("OPSORCH_TLS_KEY_FILE"))

	if (tlsCertFile == "") != (tlsKeyFile == "") {
		return nil, fmt.Errorf("both OPSORCH_TLS_CERT_FILE and OPSORCH_TLS_KEY_FILE must be set together")
	}

	sec, err := newSecretProviderFromEnv()
	if err != nil {
		return nil, err
	}

	inc, err := newIncidentHandlerFromEnv(sec)
	if err != nil {
		return nil, err
	}
	al, err := newAlertHandlerFromEnv(sec)
	if err != nil {
		return nil, err
	}
	lg, err := newLogHandlerFromEnv(sec)
	if err != nil {
		return nil, err
	}
	mt, err := newMetricHandlerFromEnv(sec)
	if err != nil {
		return nil, err
	}
	tk, err := newTicketHandlerFromEnv(sec)
	if err != nil {
		return nil, err
	}
	msg, err := newMessagingHandlerFromEnv(sec)
	if err != nil {
		return nil, err
	}
	svc, err := newServiceHandlerFromEnv(sec)
	if err != nil {
		return nil, err
	}
	dep, err := newDeploymentHandlerFromEnv(sec)
	if err != nil {
		// Log the error but continue startup with deployment capability disabled
		log.Printf("Failed to initialize deployment provider: %v", err)
		dep = DeploymentHandler{} // Empty handler with nil provider
	}

	_ = ctx // reserved for future use

	return &Server{
		corsOrigin:  corsOrigin,
		bearerToken: bearer,
		tlsCertFile: tlsCertFile,
		tlsKeyFile:  tlsKeyFile,
		incident:    inc,
		alert:       al,
		log:         lg,
		metric:      mt,
		ticket:      tk,
		messaging:   msg,
		service:     svc,
		deployment:  dep,
		secret:      sec,
	}, nil
}

// ServeHTTP implements http.Handler and dispatches to capability handlers.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers for frontend consumption.
	w.Header().Set("Access-Control-Allow-Origin", s.corsOrigin)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if !s.authorize(r) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	requestID := requestIDFromRequest(r)

	// Set headers for downstream
	w.Header().Set("X-Request-ID", requestID)

	switch {
	case r.URL.Path == "/" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case r.URL.Path == "/health" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case s.handleProviders(w, r):
	case s.handleProviderConfig(w, r):
	case s.handleIncident(w, r):
	case s.handleAlert(w, r):
	case s.handleLog(w, r):
	case s.handleMetric(w, r):
	case s.handleTicket(w, r):
	case s.handleMessaging(w, r):
	case s.handleService(w, r):
	case s.handleDeployment(w, r):
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) authorize(r *http.Request) bool {
	if s.bearerToken == "" {
		return true
	}

	const prefix = "Bearer "
	authz := r.Header.Get("Authorization")

	if !strings.HasPrefix(authz, prefix) {
		return false
	}

	token := strings.TrimSpace(authz[len(prefix):])
	return token == s.bearerToken
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{Addr: addr, Handler: s}

	serve := s.serve
	if serve == nil {
		serve = func(srv *http.Server) error { return srv.ListenAndServe() }
	}
	serveTLS := s.serveTLS
	if serveTLS == nil {
		serveTLS = func(srv *http.Server, cert, key string) error { return srv.ListenAndServeTLS(cert, key) }
	}

	// Enable TLS when both cert and key are provided.
	if s.tlsCertFile != "" || s.tlsKeyFile != "" {
		if s.tlsCertFile == "" || s.tlsKeyFile == "" {
			return fmt.Errorf("TLS requires both cert and key to be configured")
		}
		return serveTLS(srv, s.tlsCertFile, s.tlsKeyFile)
	}

	return serve(srv)
}
