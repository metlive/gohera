package okhttp

import (
	"testing"
)

func TestSetDefaultService(t *testing.T) {
	SetDefaultService("")
	t.Cleanup(func() { SetDefaultService("") })

	if DefaultService() != "" {
		t.Fatal("expected empty default")
	}
	SetDefaultService("billing")
	if DefaultService() != "billing" {
		t.Fatalf("got %q", DefaultService())
	}

	h := NewRequest().SetService("override")
	if h.resolveService() != "override" {
		t.Fatalf("got %q", h.resolveService())
	}
	h2 := NewRequest()
	if h2.resolveService() != "billing" {
		t.Fatalf("got %q", h2.resolveService())
	}
}

func TestHeaderToMap(t *testing.T) {
	m := headerToMap(map[string][]string{
		"A": {"1", "2"},
		"B": {},
	})
	if m["A"] != "1" {
		t.Fatalf("got %#v", m["A"])
	}
	if m["B"] != "" {
		t.Fatalf("got %#v", m["B"])
	}
}
