package metric

import (
	"context"
	"testing"
	"time"

	"github.com/opsorch/opsorch-core/schema"
)

type stubMetricProvider struct{}

func (stubMetricProvider) Query(ctx context.Context, q schema.MetricQuery) ([]schema.MetricSeries, error) {
	return []schema.MetricSeries{{Name: "cpu", Points: []schema.MetricPoint{{Timestamp: time.Now(), Value: 1}}}}, nil
}

func (stubMetricProvider) Describe(ctx context.Context, scope schema.QueryScope) ([]schema.MetricDescriptor, error) {
	return []schema.MetricDescriptor{{Name: "cpu", Type: "gauge"}}, nil
}

func TestMetricRegisterLookup(t *testing.T) {
	name := "test-metric"
	ctor := func(cfg map[string]any) (Provider, error) { return stubMetricProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-metric already registered" {
		t.Fatalf("register: %v", err)
	}
	_, ok := LookupProvider(name)
	if !ok {
		t.Fatalf("expected provider lookup success")
	}
	names := Providers()
	found := false
	for _, n := range names {
		if n == name {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected provider name in list: %v", names)
	}
}

func TestMetricDuplicateFails(t *testing.T) {
	name := "dup-metric"
	ctor := func(cfg map[string]any) (Provider, error) { return stubMetricProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}
