package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsorch/opsorch-core/incident"
	"github.com/opsorch/opsorch-core/log"
)

func TestProviderConfigUnknownCapability(t *testing.T) {
	mem := &memorySecret{store: map[string]string{}}
	srv := &Server{secret: mem}
	body, _ := json.Marshal(map[string]any{"provider": "stub", "config": map[string]any{"k": "v"}})
	req := httptest.NewRequest(http.MethodPost, "/providers/unknown", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProviderConfigInvalidJSON(t *testing.T) {
	mem := &memorySecret{store: map[string]string{}}
	srv := &Server{secret: mem}
	req := httptest.NewRequest(http.MethodPost, "/providers/incident", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProviderConfigUnknownProviderName(t *testing.T) {
	mem := &memorySecret{store: map[string]string{}}
	srv := &Server{secret: mem}
	body, _ := json.Marshal(map[string]any{"provider": "does-not-exist", "config": map[string]any{}})
	req := httptest.NewRequest(http.MethodPost, "/providers/incident", bytes.NewReader(body))
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProviderConfigAppliesStoredConfigOnLoad(t *testing.T) {
	name := "reg-load-stub"
	if err := incident.RegisterProvider(name, func(cfg map[string]any) (incident.Provider, error) { return stubIncidentProvider{}, nil }); err != nil && err.Error() != "registry: provider reg-load-stub already registered" {
		t.Fatalf("register incident: %v", err)
	}

	stored, _ := json.Marshal(map[string]any{"provider": name, "config": map[string]any{"token": "x"}})
	mem := &memorySecret{store: map[string]string{"providers/incident/default": string(stored)}}
	srv, err := NewServerFromEnv(rctx())
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	srv.secret = mem

	// reload incident handler from stored secret
	inc, err := newIncidentHandlerFromEnv(mem)
	if err != nil {
		t.Fatalf("load from secret: %v", err)
	}
	srv.incident = inc

	// register log as well to avoid missing provider on providers list endpoint.
	_ = log.RegisterProvider("reg-log", func(cfg map[string]any) (log.Provider, error) { return stubLogProvider{}, nil })

	if srv.incident.provider == nil {
		t.Fatalf("expected incident provider hydrated from secret store")
	}
}
