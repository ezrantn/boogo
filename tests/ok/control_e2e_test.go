package ok

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ezrantn/boogo/cmd/boogo"
)

func TestControlE2EGolden(t *testing.T) {
	srcPath := filepath.Join("../testdata", "control.bpl")
	goldenPath := filepath.Join("../testdata", "control.golden.go")

	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}

	got, err := boogo.Run(src)
	if err != nil {
		t.Fatalf("boogo.Run failed: %v", err)
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	want := string(wantBytes)

	// Normalize line endings
	got = strings.TrimSpace(got)
	want = strings.TrimSpace(want)

	if got != want {
		t.Fatalf(
			"golden mismatch\n--- WANT ---\n%s\n--- GOT ---\n%s",
			want, got,
		)
	}
}
