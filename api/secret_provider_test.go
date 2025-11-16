package api

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSecretProviderViaPlugin(t *testing.T) {
	tmp := t.TempDir()
	pluginPath := filepath.Join(tmp, "secretmock")
	build := exec.Command("go", "build", "-o", pluginPath, "../plugins/secretmock")
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(tmp, "gocache"), "GOMODCACHE="+filepath.Join(tmp, "gomodcache"), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build secret plugin: %v output=%s", err, string(out))
	}

	t.Setenv("OPSORCH_SECRET_PLUGIN", pluginPath)

	prov, err := newSecretProviderFromEnv()
	if err != nil {
		t.Fatalf("load secret provider: %v", err)
	}
	if prov == nil {
		t.Fatalf("expected secret provider when plugin path set")
	}

	if err := prov.Put(context.Background(), "foo", "bar"); err != nil {
		t.Fatalf("put: %v", err)
	}

	got, err := prov.Get(context.Background(), "foo")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "bar" {
		t.Fatalf("unexpected value %q", got)
	}
}
