package appsource

import "testing"

func TestNormalizeURLsSingle(t *testing.T) {
	urls, err := NormalizeURLs("https://example.com/appstore.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 1 || urls[0] != "https://example.com/appstore.json" {
		t.Fatalf("got %#v", urls)
	}
}

func TestNormalizeURLsJSON(t *testing.T) {
	urls, err := NormalizeURLs(`{"sourceURL":"https://a.example/s","sources":["https://b.example/s"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 {
		t.Fatalf("got %#v", urls)
	}
}

func TestNormalizeURLsEmpty(t *testing.T) {
	urls, err := NormalizeURLs("  ")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 0 {
		t.Fatalf("got %#v", urls)
	}
}

func TestNormalizeURLsRejectsNonHTTP(t *testing.T) {
	if _, err := NormalizeURLs("ftp://example.com/s"); err == nil {
		t.Fatal("expected error")
	}
}
