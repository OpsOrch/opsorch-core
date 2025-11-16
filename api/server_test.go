package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/opsorch/opsorch-core/incident"
	"github.com/opsorch/opsorch-core/log"
	"github.com/opsorch/opsorch-core/schema"
	"github.com/opsorch/opsorch-core/service"
)

// stubIncidentProvider implements incident.Provider for tests.
type stubIncidentProvider struct{}

func (s stubIncidentProvider) Query(ctx context.Context, query schema.IncidentQuery) ([]schema.Incident, error) {
	res := []schema.Incident{{ID: "1", Title: "test", CreatedAt: time.Now(), UpdatedAt: time.Now()}}
	if query.Limit > 0 && query.Limit < len(res) {
		return res[:query.Limit], nil
	}
	return res, nil
}

func (s stubIncidentProvider) List(ctx context.Context) ([]schema.Incident, error) {
	return []schema.Incident{{ID: "1", Title: "test", CreatedAt: time.Now(), UpdatedAt: time.Now()}}, nil
}

func (s stubIncidentProvider) Get(ctx context.Context, id string) (schema.Incident, error) {
	return schema.Incident{ID: id, Title: "test", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubIncidentProvider) Create(ctx context.Context, in schema.CreateIncidentInput) (schema.Incident, error) {
	return schema.Incident{ID: "new", Title: in.Title, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubIncidentProvider) Update(ctx context.Context, id string, in schema.UpdateIncidentInput) (schema.Incident, error) {
	return schema.Incident{ID: id, Title: derefString(in.Title, "test"), CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s stubIncidentProvider) GetTimeline(ctx context.Context, id string) ([]schema.TimelineEntry, error) {
	return []schema.TimelineEntry{{ID: "t1", IncidentID: id, Body: "entry", At: time.Now()}}, nil
}

func (s stubIncidentProvider) AppendTimeline(ctx context.Context, id string, entry schema.TimelineAppendInput) error {
	return nil
}

func derefString(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

type stubLogProvider struct{}

func (s stubLogProvider) Query(ctx context.Context, q schema.LogQuery) ([]schema.LogEntry, error) {
	return []schema.LogEntry{{Timestamp: time.Now(), Message: "log-entry"}}, nil
}

type stubMetricProvider struct{}

func (s stubMetricProvider) Query(ctx context.Context, q schema.MetricQuery) ([]schema.MetricSeries, error) {
	return []schema.MetricSeries{{
		Name:   "cpu",
		Points: []schema.MetricPoint{{Timestamp: time.Now(), Value: 1.0}},
	}}, nil
}

type stubTicketProvider struct{}

func (s stubTicketProvider) Query(ctx context.Context, query schema.TicketQuery) ([]schema.Ticket, error) {
	now := time.Now()
	return []schema.Ticket{{ID: "t1", Title: "ticket", Status: "open", CreatedAt: now, UpdatedAt: now}}, nil
}

func (s stubTicketProvider) Get(ctx context.Context, id string) (schema.Ticket, error) {
	now := time.Now()
	return schema.Ticket{ID: id, Title: "ticket", Status: "open", CreatedAt: now, UpdatedAt: now}, nil
}

func (s stubTicketProvider) Create(ctx context.Context, in schema.CreateTicketInput) (schema.Ticket, error) {
	now := time.Now()
	return schema.Ticket{ID: "new", Title: in.Title, Status: "open", CreatedAt: now, UpdatedAt: now}, nil
}

func (s stubTicketProvider) Update(ctx context.Context, id string, in schema.UpdateTicketInput) (schema.Ticket, error) {
	now := time.Now()
	title := derefString(in.Title, "ticket")
	status := derefString(in.Status, "open")
	return schema.Ticket{ID: id, Title: title, Status: status, CreatedAt: now, UpdatedAt: now}, nil
}

type stubMessagingProvider struct{}

func (s stubMessagingProvider) Send(ctx context.Context, msg schema.Message) (schema.MessageResult, error) {
	return schema.MessageResult{ID: "m1", Channel: msg.Channel, Metadata: msg.Metadata, SentAt: time.Now()}, nil
}

type stubServiceProvider struct{}

func (s stubServiceProvider) Query(ctx context.Context, q schema.ServiceQuery) ([]schema.Service, error) {
	return []schema.Service{{ID: "svc1", Name: "Service 1"}}, nil
}

type memorySecret struct {
	store map[string]string
}

func (m *memorySecret) Get(ctx context.Context, key string) (string, error) {
	if m.store == nil {
		return "", fmt.Errorf("not found")
	}
	v, ok := m.store[key]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return v, nil
}

func (m *memorySecret) Put(ctx context.Context, key, value string) error {
	if m.store == nil {
		m.store = make(map[string]string)
	}
	m.store[key] = value
	return nil
}

func TestHealthAndCors(t *testing.T) {
	srv := &Server{corsOrigin: "*"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if status := w.Result().StatusCode; status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected CORS header '*', got %q", got)
	}
}

func TestBearerAuthRequired(t *testing.T) {
	srv := &Server{bearerToken: "secret"}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if status := w.Result().StatusCode; status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", status)
	}
}

func TestBearerAuthSuccess(t *testing.T) {
	srv := &Server{bearerToken: "secret"}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if status := w.Result().StatusCode; status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
}

func TestListenAndServeRequiresTLSPair(t *testing.T) {
	srv := &Server{tlsCertFile: "/tmp/cert"}

	if err := srv.ListenAndServe(":0"); err == nil || !strings.Contains(err.Error(), "TLS requires both cert and key") {
		t.Fatalf("expected TLS pair error, got %v", err)
	}
}

func TestListenAndServeUsesTLSWhenConfigured(t *testing.T) {
	calledTLS := false
	calledPlain := false
	srv := &Server{
		tlsCertFile: "cert.pem",
		tlsKeyFile:  "key.pem",
		serve: func(*http.Server) error {
			calledPlain = true
			return nil
		},
		serveTLS: func(_ *http.Server, cert, key string) error {
			calledTLS = true
			if cert != "cert.pem" || key != "key.pem" {
				t.Fatalf("expected cert/key to pass through, got %s/%s", cert, key)
			}
			return nil
		},
	}

	if err := srv.ListenAndServe(":443"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calledPlain {
		t.Fatalf("expected TLS serve path to be used")
	}
	if !calledTLS {
		t.Fatalf("expected TLS serve path to be invoked")
	}
}

func TestNewServerFromEnvTLSPairValidation(t *testing.T) {
	t.Setenv("OPSORCH_TLS_CERT_FILE", "/tmp/cert")
	t.Setenv("OPSORCH_TLS_KEY_FILE", "")

	// Ensure other providers do not interfere with TLS validation.
	t.Setenv("OPSORCH_INCIDENT_PROVIDER", "")
	t.Setenv("OPSORCH_LOG_PROVIDER", "")
	t.Setenv("OPSORCH_METRIC_PROVIDER", "")
	t.Setenv("OPSORCH_TICKET_PROVIDER", "")
	t.Setenv("OPSORCH_MESSAGING_PROVIDER", "")
	t.Setenv("OPSORCH_SERVICE_PROVIDER", "")

	if srv, err := NewServerFromEnv(context.Background()); err == nil {
		t.Fatalf("expected TLS validation error, got server %+v", srv)
	}
}

func TestIncidentMissingProvider(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/incidents", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if status := w.Result().StatusCode; status != http.StatusNotImplemented {
		t.Fatalf("expected 501 when provider missing, got %d", status)
	}
}

func TestIncidentList(t *testing.T) {
	srv := &Server{incident: IncidentHandler{provider: stubIncidentProvider{}}, corsOrigin: "*"}
	req := httptest.NewRequest(http.MethodGet, "/incidents", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var out []schema.Incident
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(out) != 1 || out[0].ID != "1" {
		t.Fatalf("unexpected incident response: %+v", out)
	}
}

func TestIncidentQuery(t *testing.T) {
	srv := &Server{incident: IncidentHandler{provider: stubIncidentProvider{}}, corsOrigin: "*"}
	body, _ := json.Marshal(schema.IncidentQuery{Limit: 1})
	req := httptest.NewRequest(http.MethodPost, "/incidents/query", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var out []schema.Incident
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(out) != 1 || out[0].ID != "1" {
		t.Fatalf("unexpected incident response: %+v", out)
	}
}

func TestIncidentListViaPlugin(t *testing.T) {
	tmp := t.TempDir()
	pluginPath := filepath.Join(tmp, "incidentmock")
	build := exec.Command("go", "build", "-o", pluginPath, "../plugins/incidentmock")
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(tmp, "gocache"), "GOMODCACHE="+filepath.Join(tmp, "gomodcache"), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v output=%s", err, string(out))
	}

	srv := &Server{incident: IncidentHandler{provider: newIncidentPluginProvider(pluginPath, nil)}, corsOrigin: "*"}
	req := httptest.NewRequest(http.MethodGet, "/incidents", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out []schema.Incident
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 || out[0].ID != "p1" {
		t.Fatalf("unexpected plugin incident response: %+v", out)
	}
}

func TestIncidentQueryViaPlugin(t *testing.T) {
	tmp := t.TempDir()
	pluginPath := filepath.Join(tmp, "incidentmock")
	build := exec.Command("go", "build", "-o", pluginPath, "../plugins/incidentmock")
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(tmp, "gocache"), "GOMODCACHE="+filepath.Join(tmp, "gomodcache"), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v output=%s", err, string(out))
	}

	srv := &Server{incident: IncidentHandler{provider: newIncidentPluginProvider(pluginPath, nil)}, corsOrigin: "*"}
	body, _ := json.Marshal(schema.IncidentQuery{Limit: 1})
	req := httptest.NewRequest(http.MethodPost, "/incidents/query", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out []schema.Incident
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) == 0 || out[0].ID != "p1" {
		t.Fatalf("unexpected plugin incident response: %+v", out)
	}
}

func TestLogQueryViaPlugin(t *testing.T) {
	tmp := t.TempDir()
	pluginPath := filepath.Join(tmp, "logmock")
	build := exec.Command("go", "build", "-o", pluginPath, "../plugins/logmock")
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(tmp, "gocache"), "GOMODCACHE="+filepath.Join(tmp, "gomodcache"), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build log plugin: %v output=%s", err, string(out))
	}

	srv := &Server{log: LogHandler{provider: newLogPluginProvider(pluginPath, nil)}, corsOrigin: "*"}
	body, _ := json.Marshal(schema.LogQuery{Query: "test", Start: time.Now(), End: time.Now()})
	req := httptest.NewRequest(http.MethodPost, "/logs/query", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out []schema.LogEntry
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) == 0 || !strings.Contains(out[0].Message, "plugin log") {
		t.Fatalf("unexpected log plugin response: %+v", out)
	}
}

func TestProvidersUnknownCapability(t *testing.T) {
	srv := &Server{corsOrigin: "*"}
	req := httptest.NewRequest(http.MethodGet, "/providers/unknown", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if status := w.Result().StatusCode; status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", status)
	}
}

func TestIncidentHandlerFromEnvInvalidJSON(t *testing.T) {
	name := "test-inc-from-env"
	if err := incident.RegisterProvider(name, func(cfg map[string]any) (incident.Provider, error) { return stubIncidentProvider{}, nil }); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	t.Setenv("OPSORCH_INCIDENT_PROVIDER", name)
	t.Setenv("OPSORCH_INCIDENT_CONFIG", "not-json")

	if _, err := newIncidentHandlerFromEnv(nil); err == nil {
		t.Fatalf("expected error for invalid JSON config")
	}
}

func TestIncidentTimelineAppend(t *testing.T) {
	srv := &Server{incident: IncidentHandler{provider: stubIncidentProvider{}}}
	payload := []byte(`{"kind":"note","body":"hi"}`)
	req := httptest.NewRequest(http.MethodPost, "/incidents/abc/timeline", bytes.NewReader(payload))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestProviderConfigEndpointStoresAndApplies(t *testing.T) {
	// Register stub provider constructor for incident.
	name := "stub-http"
	if err := incident.RegisterProvider(name, func(cfg map[string]any) (incident.Provider, error) { return stubIncidentProvider{}, nil }); err != nil && !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("register provider: %v", err)
	}

	mem := &memorySecret{store: map[string]string{}}
	srv := &Server{secret: mem}

	body, _ := json.Marshal(map[string]any{"provider": name, "config": map[string]any{"token": "x"}})
	req := httptest.NewRequest(http.MethodPost, "/providers/incident", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if srv.incident.provider == nil {
		t.Fatalf("expected incident provider to be set")
	}
	if _, ok := mem.store["providers/incident/default"]; !ok {
		t.Fatalf("expected config stored in secret provider")
	}
}

func TestProviderConfigRequiresSecretProvider(t *testing.T) {
	srv := &Server{}
	body, _ := json.Marshal(map[string]any{"provider": "x", "config": map[string]any{"k": "v"}})
	req := httptest.NewRequest(http.MethodPost, "/providers/log", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

func TestLogQuery(t *testing.T) {
	srv := &Server{log: LogHandler{provider: stubLogProvider{}}}
	body, _ := json.Marshal(schema.LogQuery{Query: "test", Start: time.Now(), End: time.Now()})
	req := httptest.NewRequest(http.MethodPost, "/logs/query", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out []schema.LogEntry
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 || out[0].Message != "log-entry" {
		t.Fatalf("unexpected log response: %+v", out)
	}
}

func TestMetricQuery(t *testing.T) {
	srv := &Server{metric: MetricHandler{provider: stubMetricProvider{}}}
	body, _ := json.Marshal(schema.MetricQuery{Expression: "up", Start: time.Now(), End: time.Now(), Step: time.Second})
	req := httptest.NewRequest(http.MethodPost, "/metrics/query", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out []schema.MetricSeries
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 || out[0].Name != "cpu" {
		t.Fatalf("unexpected metric response: %+v", out)
	}
}

func TestServiceQuery(t *testing.T) {
	srv := &Server{service: ServiceHandler{provider: stubServiceProvider{}}}

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var services []schema.Service
	if err := json.NewDecoder(w.Body).Decode(&services); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(services) != 1 || services[0].ID != "svc1" {
		t.Fatalf("unexpected service response: %+v", services)
	}

	body, _ := json.Marshal(schema.ServiceQuery{Name: "service"})
	req = httptest.NewRequest(http.MethodPost, "/services/query", bytes.NewReader(body))
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestTicketCreateAndGet(t *testing.T) {
	srv := &Server{ticket: TicketHandler{provider: stubTicketProvider{}}}

	bodyQuery, _ := json.Marshal(schema.TicketQuery{})
	reqList := httptest.NewRequest(http.MethodPost, "/tickets/query", bytes.NewReader(bodyQuery))
	wList := httptest.NewRecorder()
	srv.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d", wList.Code)
	}
	var tickets []schema.Ticket
	if err := json.NewDecoder(wList.Body).Decode(&tickets); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(tickets) != 1 || tickets[0].ID != "t1" {
		t.Fatalf("unexpected ticket list: %+v", tickets)
	}

	body, _ := json.Marshal(schema.CreateTicketInput{Title: "t"})
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create expected 201, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/tickets/abc", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("get expected 200, got %d", w2.Code)
	}
	var tkt schema.Ticket
	if err := json.NewDecoder(w2.Body).Decode(&tkt); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tkt.ID != "abc" {
		t.Fatalf("unexpected ticket id: %s", tkt.ID)
	}
}

func TestMessagingSend(t *testing.T) {
	srv := &Server{messaging: MessagingHandler{provider: stubMessagingProvider{}}}
	body, _ := json.Marshal(schema.Message{Channel: "c", Body: "hi", Metadata: map[string]any{"source": "stub"}})
	req := httptest.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var res schema.MessageResult
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.ID != "m1" || res.Channel != "c" {
		t.Fatalf("unexpected messaging response: %+v", res)
	}
}

func TestMissingProvidersReturn501(t *testing.T) {
	cases := []struct {
		name   string
		server *Server
		method string
		url    string
	}{
		{"log", &Server{log: LogHandler{}}, http.MethodPost, "/logs/query"},
		{"metric", &Server{metric: MetricHandler{}}, http.MethodPost, "/metrics/query"},
		{"ticket", &Server{ticket: TicketHandler{}}, http.MethodPost, "/tickets"},
		{"messaging", &Server{messaging: MessagingHandler{}}, http.MethodPost, "/messages/send"},
		{"service", &Server{service: ServiceHandler{}}, http.MethodGet, "/services"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.url, nil)
		w := httptest.NewRecorder()
		tc.server.ServeHTTP(w, req)
		if w.Code != http.StatusNotImplemented {
			t.Fatalf("%s expected 501, got %d", tc.name, w.Code)
		}
	}
}

func TestProvidersListIncludesRegistered(t *testing.T) {
	// register a unique log provider name
	name := "test-log-listing"
	if err := log.RegisterProvider(name, func(cfg map[string]any) (log.Provider, error) { return stubLogProvider{}, nil }); err != nil && !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("register log provider: %v", err)
	}

	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/providers/log", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out map[string][]string
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	providers := out["providers"]
	found := false
	for _, p := range providers {
		if strings.EqualFold(p, name) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected provider %s in list, got %v", name, providers)
	}
}

func TestProvidersListIncludesService(t *testing.T) {
	name := "test-service-listing"
	if err := service.RegisterProvider(name, func(cfg map[string]any) (service.Provider, error) { return stubServiceProvider{}, nil }); err != nil && !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("register service provider: %v", err)
	}

	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/providers/service", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out map[string][]string
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	providers := out["providers"]
	found := false
	for _, p := range providers {
		if strings.EqualFold(p, name) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected provider %s in list, got %v", name, providers)
	}
}
