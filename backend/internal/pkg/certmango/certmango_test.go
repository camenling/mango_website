package certmango

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testKey = "ZdzzTckgNQpSHnuR"

func TestEncryptDecryptRoundtrip(t *testing.T) {
	udid := "00008120-000E589211F3C01E"
	password := "abc123"
	uuid := "3C4DC7D2-E475-3375-489C-0BB8D737A653"
	zipData, err := packCertZip(udid, []byte("p12-bytes"), []byte("mp-bytes"))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := encrypt(zipData, udid, password, uuid, testKey)
	if err != nil {
		t.Fatal(err)
	}
	gotUDID, gotPass, gotUUID, gotZip, err := decrypt(blob, testKey, uuid)
	if err != nil {
		t.Fatal(err)
	}
	if gotUDID != udid || gotPass != password || gotUUID != uuid {
		t.Fatalf("fields mismatch: %s %s %s", gotUDID, gotPass, gotUUID)
	}
	if _, _, _, _, err := decrypt(blob, testKey, strings.ToUpper(uuid)); err != nil {
		t.Fatal("uuid compare should ignore case")
	}
	if !bytes.Equal(gotZip, zipData) {
		t.Fatal("zip mismatch")
	}
	names := zipNames(t, gotZip)
	if names[0] != udid+".p12" || names[1] != udid+".mobileprovision" {
		t.Fatalf("zip names: %v", names)
	}
}

func TestParseProvisionSingleUDID(t *testing.T) {
	raw := []byte(`
		<key>UUID</key>
		<string>aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee</string>
		<key>ProvisionedDevices</key>
		<array>
			<string>00008140-001A59E10EFB001C</string>
		</array>
	`)
	uuid, udid, err := parseProvision(raw)
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Fatalf("uuid=%s", uuid)
	}
	if udid != "00008140-001A59E10EFB001C" {
		t.Fatalf("udid=%s", udid)
	}
}

func TestParseProvisionMultipleUsesPlaceholder(t *testing.T) {
	raw := []byte(`
		<key>UUID</key>
		<string>aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee</string>
		<key>ProvisionedDevices</key>
		<array>
			<string>00008140-001A59E10EFB001C</string>
			<string>00008120-000E589211F3C01E</string>
		</array>
	`)
	_, udid, err := parseProvision(raw)
	if err != nil {
		t.Fatal(err)
	}
	if udid != PlaceholderUDID {
		t.Fatalf("udid=%s", udid)
	}
}

func TestParseProvisionMissingUDIDUsesPlaceholder(t *testing.T) {
	_, udid, err := parseProvision([]byte(`<key>UUID</key><string>aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee</string>`))
	if err != nil {
		t.Fatal(err)
	}
	if udid != PlaceholderUDID {
		t.Fatalf("udid=%s", udid)
	}
}

func TestParseProvisionReadsPlistInsideCMSWrapper(t *testing.T) {
	raw := []byte("\x30\x82fake-cms-header<?xml version=\"1.0\"?><plist><dict>" +
		"<key>UUID</key><string>aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee</string>" +
		"<key>ProvisionedDevices</key><array><string>00008140-001A59E10EFB001C</string></array>" +
		"</dict></plist>\x00trailer")
	uuid, udid, err := parseProvision(raw)
	if err != nil {
		t.Fatal(err)
	}
	if uuid != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Fatalf("uuid=%s", uuid)
	}
	if udid != "00008140-001A59E10EFB001C" {
		t.Fatalf("udid=%s", udid)
	}
}

func TestParseProvisionEnterpriseUsesPlaceholder(t *testing.T) {
	raw := []byte(`
		<key>UUID</key>
		<string>aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee</string>
		<key>ProvisionsAllDevices</key>
		<true/>
	`)
	_, udid, err := parseProvision(raw)
	if err != nil {
		t.Fatal(err)
	}
	if udid != PlaceholderUDID {
		t.Fatalf("udid=%s", udid)
	}
}

func TestEmbedWritesCertMango(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "Payload", "App.app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	p12Path := filepath.Join(dir, "cert.p12")
	mpPath := filepath.Join(dir, "profile.mobileprovision")
	if err := os.WriteFile(p12Path, []byte("p12-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	mp := []byte(`
		<key>UUID</key>
		<string>aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee</string>
		<key>ProvisionedDevices</key>
		<array>
			<string>00008140-001A59E10EFB001C</string>
		</array>
	`)
	if err := os.WriteFile(mpPath, mp, 0o644); err != nil {
		t.Fatal(err)
	}

	blob, err := encryptMust(p12Path, mpPath, "pw1")
	if err != nil {
		t.Fatal(err)
	}
	udid, pass, uuid, zipData, err := decrypt(blob, testKey, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	if err != nil {
		t.Fatal(err)
	}
	if udid != "00008140-001A59E10EFB001C" || pass != "pw1" || uuid != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Fatalf("got %s %s %s", udid, pass, uuid)
	}
	if len(zipData) == 0 {
		t.Fatal("empty zip")
	}
}

func encryptMust(p12Path, mpPath, password string) ([]byte, error) {
	mp := mustRead(mpPath)
	uuid, udid, err := parseProvision(mp)
	if err != nil {
		return nil, err
	}
	zipData, err := packCertZip(udid, mustRead(p12Path), mp)
	if err != nil {
		return nil, err
	}
	return encrypt(zipData, udid, password, uuid, testKey)
}

func mustRead(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}

func zipNames(t *testing.T, data []byte) []string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, 0, len(r.File))
	for _, f := range r.File {
		out = append(out, f.Name)
	}
	return out
}
