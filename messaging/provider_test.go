package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/opsorch/opsorch-core/schema"
)

type stubMessagingProvider struct{}

func (stubMessagingProvider) Send(ctx context.Context, msg schema.Message) (schema.MessageResult, error) {
	return schema.MessageResult{ID: "m1", Channel: msg.Channel, Metadata: msg.Metadata, SentAt: time.Now()}, nil
}

func TestMessagingRegisterLookup(t *testing.T) {
	name := "test-messaging"
	ctor := func(cfg map[string]any) (Provider, error) { return stubMessagingProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-messaging already registered" {
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

func TestMessagingDuplicateFails(t *testing.T) {
	name := "dup-messaging"
	ctor := func(cfg map[string]any) (Provider, error) { return stubMessagingProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}
