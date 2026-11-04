package certzip_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"mangosign/internal/pkg/certzip"
)

func TestExtractZip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "certs.zip")
	createTestZip(t, zipPath)

	extractDir := filepath.Join(dir, "extracted")
	files, err := certzip.ExtractZip(zipPath, extractDir)
	if err != nil {
		t.Fatal(err)
	}
	if files.P12Name != "demo.p12" {
		t.Fatalf("unexpected p12: %s", files.P12Name)
	}
	if files.ProvisionName != "demo.mobileprovision" {
		t.Fatalf("unexpected provision: %s", files.ProvisionName)
	}
}

func createTestZip(t *testing.T, zipPath string) {
	t.Helper()
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range map[string]string{
		"folder/demo.p12":            "p12-content",
		"folder/demo.mobileprovision": "provision-content",
	} {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
}
