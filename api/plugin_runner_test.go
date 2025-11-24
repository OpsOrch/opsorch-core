package api

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/opsorch/opsorch-core/schema"
)

// Ensure the plugin runner keeps the plugin process alive across requests even when
// individual request contexts are canceled.
func TestPluginRunnerSurvivesContextCancel(t *testing.T) {
	tmp := t.TempDir()
	pluginPath := filepath.Join(tmp, "incidentmock")
	build := exec.Command("go", "build", "-o", pluginPath, "../plugins/incidentmock")
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(tmp, "gocache"), "GOMODCACHE="+filepath.Join(tmp, "gomodcache"), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v output=%s", err, string(out))
	}

	runner := newPluginRunner(pluginPath, nil)

	ctx1, cancel1 := context.WithCancel(context.Background())
	var res1 []schema.Incident
	if err := runner.call(ctx1, "incident.query", schema.IncidentQuery{}, &res1); err != nil {
		t.Fatalf("first call: %v", err)
	}
	cancel1()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	var res2 []schema.Incident
	if err := runner.call(ctx2, "incident.query", schema.IncidentQuery{}, &res2); err != nil {
		t.Fatalf("second call after cancellation: %v", err)
	}
	if len(res2) == 0 || res2[0].ID != "p1" {
		t.Fatalf("unexpected plugin response: %+v", res2)
	}
}
