package recorder

import "testing"

func TestProcess_SingleRelativeCD_CanonicalizesToCurrentDir(t *testing.T) {
	out := Process(Options{
		Commands:   []string{"cd Desktop"},
		Count:      1,
		CurrentDir: "/Users/test/Desktop",
	})

	if len(out) != 1 {
		t.Fatalf("expected 1 command, got %d", len(out))
	}
	if out[0] != "cd '/Users/test/Desktop'" {
		t.Fatalf("expected canonical absolute cd, got %q", out[0])
	}
}

func TestProcess_SingleRelativeCD_IgnorePathSkipsCD(t *testing.T) {
	out := Process(Options{
		Commands:   []string{"cd Desktop"},
		Count:      1,
		IgnorePath: true,
		CurrentDir: "/Users/test/Desktop",
	})

	if len(out) != 0 {
		t.Fatalf("expected no commands when ignore path is enabled, got %v", out)
	}
}
