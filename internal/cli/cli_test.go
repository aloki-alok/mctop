package cli

import "testing"

func TestExtractConn(t *testing.T) {
	opts, rest, err := extractConn([]string{"--sse", "http://x/mcp", "-H", "Authorization: Bearer t"})
	if err != nil {
		t.Fatalf("extractConn: %v", err)
	}
	if !opts.SSE {
		t.Error("SSE flag not picked up")
	}
	if got := opts.Headers["Authorization"]; got != "Bearer t" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer t")
	}
	if len(rest) != 1 || rest[0] != "http://x/mcp" {
		t.Errorf("rest = %v, want [http://x/mcp]", rest)
	}
}

func TestExtractConnDefaults(t *testing.T) {
	opts, rest, err := extractConn([]string{"uvx", "mcp-server-time"})
	if err != nil {
		t.Fatalf("extractConn: %v", err)
	}
	if opts.SSE {
		t.Error("SSE should default off")
	}
	if len(rest) != 2 {
		t.Errorf("rest = %v, want both positional args kept", rest)
	}
}

func TestExtractConnHeaderNeedsValue(t *testing.T) {
	if _, _, err := extractConn([]string{"http://x", "-H"}); err == nil {
		t.Error("expected an error when -H has no argument")
	}
}
