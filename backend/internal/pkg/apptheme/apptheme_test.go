package apptheme

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func TestBuildThemePackage(t *testing.T) {
	data, err := BuildThemePackage(Config{Theme: 4, Accent: NormalizeAccent(4, "ff375f"), Glass: true})
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(zr.File) != 1 || zr.File[0].Name != "manifest.json" {
		t.Fatalf("unexpected entries: %v", zr.File)
	}
	rc, _ := zr.File[0].Open()
	body, _ := io.ReadAll(rc)
	var m struct {
		SchemaVersion int `json:"schemaVersion"`
		Modules       struct {
			Style struct {
				V     int  `json:"v"`
				Style int  `json:"style"`
				Glass bool `json:"glass"`
			} `json:"style"`
			TextColor struct {
				V      int               `json:"v"`
				Colors map[string]string `json:"colors"`
			} `json:"textColor"`
		} `json:"modules"`
	}
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	if m.SchemaVersion != 1 || m.Modules.Style.V != 1 || m.Modules.Style.Style != 4 || !m.Modules.Style.Glass ||
		m.Modules.TextColor.V != 1 || m.Modules.TextColor.Colors["accentLight"] != "#FF375F" {
		t.Fatalf("unexpected manifest: %s", body)
	}
}

func TestNormalizeAccent(t *testing.T) {
	if got := NormalizeAccent(2, "bad"); got != "#D97A55" {
		t.Fatalf("fallback: %s", got)
	}
	if got := NormalizeAccent(1, "#12abEF"); got != "#12ABEF" {
		t.Fatalf("normalize: %s", got)
	}
}

func TestValidatePackage(t *testing.T) {
	good, err := BuildThemePackage(Config{Theme: 2, Accent: NormalizeAccent(2, ""), Glass: false})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidatePackage(good); err != nil {
		t.Fatalf("valid package rejected: %v", err)
	}
	if err := ValidatePackage([]byte("not a zip")); err == nil {
		t.Fatal("garbage accepted")
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("readme.txt")
	_, _ = w.Write([]byte("x"))
	_ = zw.Close()
	if err := ValidatePackage(buf.Bytes()); err == nil {
		t.Fatal("zip without manifest accepted")
	}
	if err := ValidatePackage(make([]byte, MaxPackageBytes+1)); err == nil {
		t.Fatal("oversized package accepted")
	}
}
