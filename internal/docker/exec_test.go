package docker

import (
	"context"
	"strings"
	"testing"
)

// These tests need no Docker daemon: they cover Exec's argument validation,
// which must fail before any process is spawned.

func TestExecRejectsEmptyService(t *testing.T) {
	c := New("docker-compose.yml", t.TempDir(), "proj")
	for _, svc := range []string{"", "   ", "\t"} {
		out, err := c.Exec(context.Background(), svc, "echo", "hi")
		if err == nil {
			t.Fatalf("Exec(service=%q) must fail validation", svc)
		}
		if out != "" {
			t.Errorf("Exec must return no output on validation failure, got %q", out)
		}
		if !strings.Contains(err.Error(), "no service given") {
			t.Errorf("Exec(service=%q) error = %v, want a no-service validation error", svc, err)
		}
	}
}

func TestExecRejectsEmptyCommand(t *testing.T) {
	c := New("docker-compose.yml", t.TempDir(), "proj")
	out, err := c.Exec(context.Background(), "api")
	if err == nil {
		t.Fatal("Exec with no argv must fail validation")
	}
	if out != "" {
		t.Errorf("Exec must return no output on validation failure, got %q", out)
	}
	if !strings.Contains(err.Error(), "no command given") {
		t.Errorf("Exec(no argv) error = %v, want a no-command validation error", err)
	}
}
