package controller

import "testing"

func TestChannelIpaFileName(t *testing.T) {
	got := channelIpaFileName("HuaYang", "Release", "260906")
	if got != "mango-huayang-release-260906.ipa" {
		t.Fatalf("got %s", got)
	}
	got = channelIpaFileName("huayang", "beta", "")
	if got != "mango-huayang-beta-latest.ipa" {
		t.Fatalf("got %s", got)
	}
}
