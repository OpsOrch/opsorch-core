package ticket

import (
	"context"
	"testing"
	"time"

	"github.com/opsorch/opsorch-core/schema"
)

type stubTicketProvider struct{}

func (stubTicketProvider) Query(ctx context.Context, query schema.TicketQuery) ([]schema.Ticket, error) {
	now := time.Now()
	return []schema.Ticket{{ID: "t1", Title: "t", Status: "open", CreatedAt: now, UpdatedAt: now}}, nil
}

func (stubTicketProvider) Get(ctx context.Context, id string) (schema.Ticket, error) {
	now := time.Now()
	return schema.Ticket{ID: id, Title: "t", Status: "open", CreatedAt: now, UpdatedAt: now}, nil
}
func (stubTicketProvider) Create(ctx context.Context, in schema.CreateTicketInput) (schema.Ticket, error) {
	now := time.Now()
	return schema.Ticket{ID: "new", Title: in.Title, Status: "open", CreatedAt: now, UpdatedAt: now}, nil
}
func (stubTicketProvider) Update(ctx context.Context, id string, in schema.UpdateTicketInput) (schema.Ticket, error) {
	now := time.Now()
	return schema.Ticket{ID: id, Title: "t", Status: "open", CreatedAt: now, UpdatedAt: now}, nil
}

func TestTicketRegisterLookup(t *testing.T) {
	name := "test-ticket"
	ctor := func(cfg map[string]any) (Provider, error) { return stubTicketProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-ticket already registered" {
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

func TestTicketDuplicateFails(t *testing.T) {
	name := "dup-ticket"
	ctor := func(cfg map[string]any) (Provider, error) { return stubTicketProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}
