package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	enterKey = tea.KeyMsg{Type: tea.KeyEnter}
	escKey   = tea.KeyMsg{Type: tea.KeyEsc}
)

func tableModel() model {
	m := model{screen: result, width: 80, height: 24, spin: newSpinner(), vim: true}
	m.vp = viewport.New(80, 10)
	m.output = `[{"name":"alpha","n":1},{"name":"beta","n":2},{"name":"gamma","n":3}]`
	m.rows = asObjectRows(decodeMust(m.output))
	m.vp.SetContent(m.resultBody())
	return m
}

func TestAsObjectRows(t *testing.T) {
	if rows := asObjectRows(decodeMust(`[{"a":1},{"b":2}]`)); len(rows) != 2 {
		t.Fatalf("array of objects should yield rows, got %d", len(rows))
	}
	for _, raw := range []string{`[1,2,3]`, `[]`, `{"a":1}`, `[{"a":1},2]`, `"x"`} {
		if asObjectRows(decodeMust(raw)) != nil {
			t.Fatalf("%s should not be rows", raw)
		}
	}
}

func TestTableMarksSelectedRow(t *testing.T) {
	rows := asObjectRows(decodeMust(`[{"a":1},{"a":2}]`))
	got, ok := renderObjectTable(rows, 80, 1)
	if !ok {
		t.Fatal("should render")
	}
	lines := strings.Split(stripANSI(got), "\n") // header, divider, row0, row1
	if strings.Contains(lines[2], "▌") {
		t.Fatalf("row 0 should not be marked:\n%s", got)
	}
	if !strings.Contains(lines[3], "▌") {
		t.Fatalf("row 1 should carry the selection marker:\n%s", got)
	}
}

func TestRenderObjectTableNoMarkerWhenUnselected(t *testing.T) {
	rows := asObjectRows(decodeMust(`[{"a":1},{"a":2}]`))
	got, _ := renderObjectTable(rows, 80, -1)
	if strings.Contains(got, "▌") {
		t.Fatal("an unselected table should have no marker gutter")
	}
}

func TestCallResultDetectsTable(t *testing.T) {
	m := model{width: 80, height: 24, spin: newSpinner()}
	m.vp = viewport.New(80, 10)
	m, _ = send(m, callResultMsg{output: `[{"a":1},{"a":2}]`, elapsed: "1ms"})
	if len(m.rows) != 2 {
		t.Fatalf("array-of-objects result should be row-navigable, got %d rows", len(m.rows))
	}
	m, _ = send(m, callResultMsg{output: `{"a":1}`, elapsed: "1ms"})
	if m.rows != nil {
		t.Fatal("a single object result should not be row-navigable")
	}
	m, _ = send(m, callResultMsg{err: errors.New("boom"), output: `[{"a":1}]`, elapsed: "1ms"})
	if m.rows != nil {
		t.Fatal("an errored result should not be row-navigable")
	}
}

func TestRowNavigationMovesSelectsAndExpands(t *testing.T) {
	m := tableModel() // three rows, vim on
	m, _ = send(m, key("down"))
	if m.rowCursor != 1 {
		t.Fatalf("down: want row 1, got %d", m.rowCursor)
	}
	m, _ = send(m, key("G"))
	if m.rowCursor != 2 {
		t.Fatalf("G: want last row, got %d", m.rowCursor)
	}
	m, _ = send(m, key("g"))
	if m.rowCursor != 0 {
		t.Fatalf("g: want first row, got %d", m.rowCursor)
	}
	m, _ = send(m, enterKey)
	if !m.rowOpen {
		t.Fatal("enter should expand the selected row")
	}
	m, _ = send(m, escKey)
	if m.rowOpen {
		t.Fatal("esc should collapse the row back to the list")
	}
	if m.screen != result {
		t.Fatal("collapsing should stay on the result screen")
	}
	m, _ = send(m, escKey)
	if m.screen != browse {
		t.Fatal("esc from the list should return to browse")
	}
}

func TestExpandedRowKeysScrollNotSelect(t *testing.T) {
	m := tableModel()
	m, _ = send(m, enterKey) // expand row 0
	m, _ = send(m, key("j")) // should scroll the detail, not move the row cursor
	if m.rowCursor != 0 {
		t.Fatalf("j in an expanded row should not move the row cursor, got %d", m.rowCursor)
	}
}

func TestRowDetailShowsFullValue(t *testing.T) {
	note := strings.Repeat("ab", 80) // 160 chars, far over the table cell cap
	m := model{screen: result, width: 80, height: 24, spin: newSpinner(), vim: true}
	m.vp = viewport.New(80, 10)
	m.output = `[{"name":"alpha","note":"` + note + `"}]`
	m.rows = asObjectRows(decodeMust(m.output))
	m.vp.SetContent(m.resultBody())

	if strings.Contains(stripANSI(m.resultBody()), note) {
		t.Fatal("the table cell should be truncated, not show the whole note")
	}
	m, _ = send(m, enterKey)
	flat := strings.NewReplacer("\n", "", " ", "").Replace(stripANSI(m.resultBody()))
	if !strings.Contains(flat, note) {
		t.Fatal("the expanded row should show the note in full")
	}
}
