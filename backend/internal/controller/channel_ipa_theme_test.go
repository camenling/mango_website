package controller

import (
	"encoding/json"
	"testing"
)

func TestSplitChannelIpaTheme(t *testing.T) {
	body := []byte(`{"v":1,"kid":"k","enc_key":"a","ciphertext":"b","tag":"c","theme":{"style":3,"accent":"#30b0c7","glass":false}}`)
	envelope, theme, err := splitChannelIpaTheme(body)
	if err != nil {
		t.Fatal(err)
	}
	if theme == nil || theme.Style != 3 || theme.Accent != "#30b0c7" || theme.Glass == nil || *theme.Glass {
		t.Fatalf("theme not parsed: %+v", theme)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(envelope, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["theme"]; ok {
		t.Fatal("theme should be stripped from the forwarded envelope")
	}
	if m["kid"] != "k" || m["tag"] != "c" {
		t.Fatalf("envelope fields lost: %s", envelope)
	}

	plain := []byte(`{"v":1,"kid":"k"}`)
	envelope, theme, err = splitChannelIpaTheme(plain)
	if err != nil || theme != nil || string(envelope) != string(plain) {
		t.Fatalf("body without theme should pass through: %s %v %v", envelope, theme, err)
	}

	if _, _, err = splitChannelIpaTheme([]byte(`{"theme":"oops"}`)); err == nil {
		t.Fatal("non-object theme accepted")
	}
}
