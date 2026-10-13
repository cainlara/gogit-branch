package execution

import (
	"reflect"
	"testing"
)

func TestBuildSelectRowsOrderAndState(t *testing.T) {
	rows := buildSelectRows([]string{"a.txt", "b.txt"})

	if len(rows) != 4 {
		t.Fatalf("expected 4 rows (Toggle all, Done, 2 files), got %d", len(rows))
	}

	if rows[0].kind != selectRowKindToggleAll || rows[1].kind != selectRowKindDone {
		t.Fatalf("sentinels must come first, got kinds %d, %d", rows[0].kind, rows[1].kind)
	}

	for _, row := range rows[2:] {
		if row.kind != selectRowKindFile {
			t.Fatalf("expected file rows after sentinels, got kind %d", row.kind)
		}

		if row.IsSelected() {
			t.Fatalf("file rows must start deselected, %q is selected", row.Label())
		}
	}
}

func TestToggleAllRowsFlipsOnlyFileRows(t *testing.T) {
	rows := buildSelectRows([]string{"a.txt", "b.txt"})

	toggleAllRows(rows)

	for _, row := range rows[2:] {
		if !row.IsSelected() {
			t.Fatalf("expected %q selected after toggle-all", row.Label())
		}
	}

	toggleAllRows(rows)

	for _, row := range rows[2:] {
		if row.IsSelected() {
			t.Fatalf("expected %q deselected after second toggle-all", row.Label())
		}
	}
}

func TestSelectedPathsPreservesListOrder(t *testing.T) {
	rows := buildSelectRows([]string{"a.txt", "b.txt", "c.txt"})

	rows[2].file.SetSelected(true)
	rows[4].file.SetSelected(true)

	got := selectedPaths(rows)
	want := []string{"a.txt", "c.txt"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selectedPaths = %v, want %v", got, want)
	}
}

func TestSelectedPathsEmptyIsCancel(t *testing.T) {
	rows := buildSelectRows([]string{"a.txt"})

	if got := selectedPaths(rows); got != nil {
		t.Fatalf("expected nil (cancel) with nothing selected, got %v", got)
	}
}
