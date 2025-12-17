package team

import (
	"context"
	"testing"

	"github.com/opsorch/opsorch-core/schema"
)

type stubTeamProvider struct{}

func (stubTeamProvider) Query(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error) {
	return []schema.Team{{ID: "team1", Name: "Test Team"}}, nil
}

func (stubTeamProvider) Get(ctx context.Context, id string) (schema.Team, error) {
	return schema.Team{ID: id, Name: "Test Team"}, nil
}

func (stubTeamProvider) Members(ctx context.Context, teamID string) ([]schema.TeamMember, error) {
	return []schema.TeamMember{{ID: "user1", Name: "Test User", Email: "test@example.com"}}, nil
}

func TestTeamRegisterLookup(t *testing.T) {
	name := "test-team"
	ctor := func(cfg map[string]any) (Provider, error) { return stubTeamProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-team already registered" {
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

func TestTeamDuplicateFails(t *testing.T) {
	name := "dup-team"
	ctor := func(cfg map[string]any) (Provider, error) { return stubTeamProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}