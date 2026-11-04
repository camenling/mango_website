package siteaes

import "testing"

func TestRoundTrip(t *testing.T) {
	plain := `{"channel":"release","bundle_id":""}`
	cipher, err := Encrypt("MangoSiteSign2026K7pQ2nR8wL4xH9vB", plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decrypt("MangoSiteSign2026K7pQ2nR8wL4xH9vB", cipher)
	if err != nil || got != plain {
		t.Fatalf("got %q err %v", got, err)
	}
}
