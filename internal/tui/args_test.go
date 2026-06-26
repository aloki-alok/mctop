package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestEnterWithEmptyRequiredFieldBlocksSubmit(t *testing.T) {
	m := model{tools: []*sdk.Tool{toolWithArgs()}, width: 80}
	opened, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = opened.(model) // form is open, timezone (required) is empty

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(model)
	if got.screen != form {
		t.Fatalf("submit with empty required field should stay on the form, screen=%v", got.screen)
	}
	if got.formMsg == "" {
		t.Error("a missing required field should set a validation message")
	}
	if cmd != nil {
		t.Error("no call should be dispatched while a required field is empty")
	}
}

func TestToolArgsReadsEnumDefaultFormat(t *testing.T) {
	tool := &sdk.Tool{
		InputSchema: map[string]any{
			"properties": map[string]any{
				"mode": map[string]any{
					"type": "string",
					"enum": []any{"fast", "slow"},
				},
				"limit": map[string]any{
					"type":    "integer",
					"default": float64(10),
				},
				"since": map[string]any{
					"type":   "string",
					"format": "date-time",
				},
			},
			"required": []any{"mode"},
		},
	}
	byName := map[string]Arg{}
	for _, a := range toolArgs(tool) {
		byName[a.Name] = a
	}

	if got := byName["mode"]; len(got.Enum) != 2 || got.Enum[0] != "fast" {
		t.Errorf("mode enum: want [fast slow], got %v", got.Enum)
	}
	if got := byName["limit"].Default; got != "10" {
		t.Errorf("limit default: want 10, got %q", got)
	}
	if got := byName["since"].Format; got != "date-time" {
		t.Errorf("since format: want date-time, got %q", got)
	}
}

func TestSearchMatchesDescription(t *testing.T) {
	m := model{tools: []*sdk.Tool{
		{Name: "alpha", Description: "convert between timezones"},
		{Name: "beta", Description: "fetch a url"},
	}}
	m.query = "between" // appears only in alpha's description
	vis := m.visibleItems()
	if len(vis) != 1 || vis[0] != 0 {
		t.Fatalf("want only alpha to match by description, got %v", vis)
	}
}

func TestArgHintPrefersEnumThenDefaultThenFormat(t *testing.T) {
	cases := []struct {
		name string
		arg  Arg
		want string
	}{
		{"enum wins", Arg{Type: "string", Enum: []string{"a", "b"}, Default: "a", Format: "x"}, "a | b"},
		{"default next", Arg{Type: "integer", Default: "10", Format: "x"}, "default 10"},
		{"format last", Arg{Type: "string", Format: "date-time"}, "date-time"},
		{"plain type has no hint", Arg{Type: "string"}, ""},
	}
	for _, c := range cases {
		if got := c.arg.hint(); got != c.want {
			t.Errorf("%s: hint()=%q, want %q", c.name, got, c.want)
		}
	}
}
