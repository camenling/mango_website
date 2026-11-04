package ipazip

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZipUnzipRoundtrip(t *testing.T) {
	src := t.TempDir()
	app := filepath.Join(src, "Payload", "Demo.app")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "Info.plist"), []byte("plist"), 0o644); err != nil {
		t.Fatal(err)
	}
	ipa := filepath.Join(t.TempDir(), "out.ipa")
	if err := ZipDir(src, ipa); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := Unzip(ipa, dest); err != nil {
		t.Fatal(err)
	}
	found, err := FindApp(dest)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(found) != "Demo.app" {
		t.Fatalf("app %s", found)
	}
	got, err := os.ReadFile(filepath.Join(found, "Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "plist" {
		t.Fatalf("plist %q", got)
	}
}

func TestUnzipRejectsZipSlip(t *testing.T) {
	ipa := filepath.Join(t.TempDir(), "slip.ipa")
	f, err := os.Create(ipa)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	fw, err := w.Create("../escape.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fw.Write([]byte("nope")); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
	dest := filepath.Join(t.TempDir(), "extract")
	if err = Unzip(ipa, dest); err == nil {
		t.Fatal("expected zip slip reject")
	}
	if _, err = os.Stat(filepath.Join(t.TempDir(), "escape.txt")); err == nil {
		t.Fatal("escaped")
	}
}

func TestUniqueWorkDirs(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	if a == b {
		t.Fatal("dirs collided")
	}
	if err := os.MkdirAll(filepath.Join(a, "Payload"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(b, "Payload"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a, "Payload", "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(b, "Payload", "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(a, "Payload", "b.txt")); err == nil {
		t.Fatal("dirs leaked")
	}
}

func TestSafeJoin(t *testing.T) {
	dest := t.TempDir()
	if _, err := safeJoin(dest, "../x"); err == nil {
		t.Fatal("expected slip")
	}
	got, err := safeJoin(dest, "Payload/App.app/Info.plist")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, dest) {
		t.Fatalf("got %s dest %s", got, dest)
	}
}
