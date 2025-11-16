package secret

import "testing"

func TestRegisterAndLookup(t *testing.T) {
	ctor := func(cfg map[string]any) (Provider, error) { return nil, nil }
	if err := RegisterProvider("vault", ctor); err != nil && err.Error() != "registry: provider vault already registered" {
		t.Fatalf("register: %v", err)
	}
	got, ok := LookupProvider("Vault")
	if !ok || got == nil {
		t.Fatalf("expected to find provider")
	}
}

func TestProvidersList(t *testing.T) {
	ctor := func(cfg map[string]any) (Provider, error) { return nil, nil }
	_ = RegisterProvider("aws-kms", ctor)
	names := Providers()
	found := false
	for _, n := range names {
		if n == "aws-kms" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected aws-kms in providers: %v", names)
	}
}

func TestEmptyNameRejected(t *testing.T) {
	ctor := func(cfg map[string]any) (Provider, error) { return nil, nil }
	if err := RegisterProvider("", ctor); err == nil {
		t.Fatalf("expected error on empty name")
	}
}
