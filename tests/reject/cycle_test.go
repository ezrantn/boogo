package reject

import (
	"os"
	"testing"

	"github.com/ezrantn/boogo/cmd/boogo"
)

func TestRejectCycleE2E(t *testing.T) {
	src, err := os.ReadFile("cycle.bpl")
	if err != nil {
		t.Fatalf("read input: %v", err)
	}

	if _, err := boogo.Run(src); err == nil {
		t.Fatalf("expected program to be rejected")
	}
}
