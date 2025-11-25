package secret

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestBasicProvider_FileConfig(t *testing.T) {
	f, err := os.CreateTemp("", "secrets-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	content := `{"secret_key": "secret_value"}`
	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg := map[string]any{
		"path": f.Name(),
	}
	p, err := NewJsonProvider(cfg)
	if err != nil {
		t.Fatalf("NewJsonProvider: %v", err)
	}

	ctx := context.Background()
	got, err := p.Get(ctx, "secret_key")
	if err != nil {
		t.Fatalf("Get(secret_key): %v", err)
	}
	if got != "secret_value" {
		t.Errorf("Get(secret_key) = %q, want %q", got, "secret_value")
	}
}

func TestBasicProvider_MissingPath(t *testing.T) {
	cfg := map[string]any{
		"foo": "bar",
	}
	_, err := NewJsonProvider(cfg)
	if err == nil {
		t.Fatal("expected error when path is missing")
	}
}

func TestBasicProvider_Registration(t *testing.T) {
	ctor, ok := LookupProvider("json")
	if !ok {
		t.Fatal("json provider not registered")
	}
	f, _ := os.CreateTemp("", "secrets-*.json")
	defer os.Remove(f.Name())
	_ = os.WriteFile(f.Name(), []byte(`{"k":"v"}`), 0600)

	p, err := ctor(map[string]any{"path": f.Name()})
	if err != nil {
		t.Fatal(err)
	}
	val, err := p.Get(context.Background(), "k")
	if err != nil {
		t.Fatal(err)
	}
	if val != "v" {
		t.Errorf("got %q, want %q", val, "v")
	}
}

func TestJsonProvider_ComplexNestedConfig(t *testing.T) {
	f, err := os.CreateTemp("", "secrets-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Complex nested config similar to PagerDuty provider
	content := `{
		"pagerduty": {
			"source": "pagerduty",
			"defaultSeverity": "critical",
			"apiURL": "https://api.pagerduty.com",
			"apiToken": "pd_live_abc123",
			"serviceID": "PXYZ123",
			"fromEmail": "ops@company.com"
		},
		"jira": {
			"apiToken": "jira_token_xyz",
			"apiURL": "https://company.atlassian.net",
			"projectKey": "OPS"
		}
	}`
	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg := map[string]any{
		"path": f.Name(),
	}
	p, err := NewJsonProvider(cfg)
	if err != nil {
		t.Fatalf("NewJsonProvider: %v", err)
	}

	ctx := context.Background()

	// Test getting pagerduty config
	pdConfigJSON, err := p.Get(ctx, "pagerduty")
	if err != nil {
		t.Fatalf("Get(pagerduty): %v", err)
	}

	// Verify it's valid JSON and can be unmarshaled
	var pdConfig map[string]any
	if err := json.Unmarshal([]byte(pdConfigJSON), &pdConfig); err != nil {
		t.Fatalf("failed to unmarshal pagerduty config: %v", err)
	}

	// Verify nested values
	if pdConfig["apiToken"] != "pd_live_abc123" {
		t.Errorf("pagerduty apiToken = %q, want %q", pdConfig["apiToken"], "pd_live_abc123")
	}
	if pdConfig["serviceID"] != "PXYZ123" {
		t.Errorf("pagerduty serviceID = %q, want %q", pdConfig["serviceID"], "PXYZ123")
	}
	if pdConfig["fromEmail"] != "ops@company.com" {
		t.Errorf("pagerduty fromEmail = %q, want %q", pdConfig["fromEmail"], "ops@company.com")
	}

	// Test getting jira config
	jiraConfigJSON, err := p.Get(ctx, "jira")
	if err != nil {
		t.Fatalf("Get(jira): %v", err)
	}

	var jiraConfig map[string]any
	if err := json.Unmarshal([]byte(jiraConfigJSON), &jiraConfig); err != nil {
		t.Fatalf("failed to unmarshal jira config: %v", err)
	}

	if jiraConfig["apiToken"] != "jira_token_xyz" {
		t.Errorf("jira apiToken = %q, want %q", jiraConfig["apiToken"], "jira_token_xyz")
	}
	if jiraConfig["projectKey"] != "OPS" {
		t.Errorf("jira projectKey = %q, want %q", jiraConfig["projectKey"], "OPS")
	}
}

func TestJsonProvider_MixedSimpleAndComplexValues(t *testing.T) {
	f, err := os.CreateTemp("", "secrets-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Mix of simple strings and complex objects
	content := `{
		"db_password": "supersecret",
		"api_key": "simple_key_123",
		"pagerduty": {
			"apiToken": "pd_token",
			"serviceID": "PD123"
		}
	}`
	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg := map[string]any{
		"path": f.Name(),
	}
	p, err := NewJsonProvider(cfg)
	if err != nil {
		t.Fatalf("NewJsonProvider: %v", err)
	}

	ctx := context.Background()

	// Test simple string value
	dbPass, err := p.Get(ctx, "db_password")
	if err != nil {
		t.Fatalf("Get(db_password): %v", err)
	}
	if dbPass != "supersecret" {
		t.Errorf("db_password = %q, want %q", dbPass, "supersecret")
	}

	// Test complex object value
	pdConfigJSON, err := p.Get(ctx, "pagerduty")
	if err != nil {
		t.Fatalf("Get(pagerduty): %v", err)
	}

	var pdConfig map[string]any
	if err := json.Unmarshal([]byte(pdConfigJSON), &pdConfig); err != nil {
		t.Fatalf("failed to unmarshal pagerduty config: %v", err)
	}

	if pdConfig["apiToken"] != "pd_token" {
		t.Errorf("pagerduty apiToken = %q, want %q", pdConfig["apiToken"], "pd_token")
	}
}
