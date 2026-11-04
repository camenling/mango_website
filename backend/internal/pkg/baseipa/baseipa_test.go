package baseipa

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadRelPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://auth.mangoios.com/upload_file/2026-08-17/dkr34qzlhot40fhenr.ipa", "upload_file/2026-08-17/dkr34qzlhot40fhenr.ipa"},
		{"http://127.0.0.1:8808/upload_file/2026-08-17/a.ipa", "upload_file/2026-08-17/a.ipa"},
		{"/upload_file/2026-08-17/a.ipa", "upload_file/2026-08-17/a.ipa"},
		{"upload_file/2026-08-17/a.ipa", "upload_file/2026-08-17/a.ipa"},
		{"https://cdn.example.com/files/foo.ipa", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := uploadRelPath(c.in); got != c.want {
			t.Fatalf("uploadRelPath(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestLookupTrackRejectsBadTrack(t *testing.T) {
	if _, err := LookupTrack(context.Background(), "nightly"); err == nil {
		t.Fatal("nightly should fail")
	}
}

func TestNormalizeUpdateTrack(t *testing.T) {
	got, err := NormalizeUpdateTrack("")
	if err != nil || got != "release" {
		t.Fatalf("empty -> release, got %q %v", got, err)
	}
	got, err = NormalizeUpdateTrack("Beta")
	if err != nil || got != "beta" {
		t.Fatalf("got %q %v", got, err)
	}
	if _, err = NormalizeUpdateTrack("nightly"); err == nil {
		t.Fatal("nightly should fail")
	}
}

func TestSanitizeName(t *testing.T) {
	if got := sanitizeName("260816"); got != "260816" {
		t.Fatalf("got %s", got)
	}
	if got := sanitizeName("26/08/16"); got != "260816" {
		t.Fatalf("got %s", got)
	}
	if got := sanitizeName("  "); got != "unknown" {
		t.Fatalf("got %s", got)
	}
}

func TestFileUsable(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.ipa")
	if fileUsable(p) {
		t.Fatal("missing file")
	}
	if err := os.WriteFile(p, []byte("ipa"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !fileUsable(p) {
		t.Fatal("expected usable")
	}
}
