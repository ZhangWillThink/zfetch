package modules

import "testing"

func TestComposeShellDisplay_fishNoise(t *testing.T) {
	raw := `fish，版本 4.7.1`
	got := ComposeShellDisplay("fish", raw)
	want := "fish 4.7.1"
	if got != want {
		t.Fatalf("ComposeShellDisplay: got %q want %q", got, want)
	}
}

func TestComposeShellDisplay_plainVersion(t *testing.T) {
	got := ComposeShellDisplay("bash", "5.2.37(1)-release")
	if got != "bash 5.2.37(1)-release" {
		t.Fatalf("got %q", got)
	}
}
