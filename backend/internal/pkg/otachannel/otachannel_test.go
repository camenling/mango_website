package otachannel

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbedChannelRoundtrip(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	app := filepath.Join(t.TempDir(), "App.app")
	if err = os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}
	kid := "abcd1234abcd1234"
	if err = EmbedChannel(context.Background(), app, "HuaYang", kid, string(pubPEM)); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(app, Rel))
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		V          int    `json:"v"`
		Kid        string `json:"kid"`
		EncKey     string `json:"enc_key"`
		Ciphertext string `json:"ciphertext"`
		Tag        string `json:"tag"`
	}
	if err = json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env.V != 1 || env.Kid != kid {
		t.Fatalf("env %+v", env)
	}
	plain, err := openTestEnvelope(&env, priv)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		HMBuildChannel string `json:"HMBuildChannel"`
	}
	if err = json.Unmarshal(plain, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.HMBuildChannel != "huayang" {
		t.Fatalf("channel %q", payload.HMBuildChannel)
	}
}

func TestEmbedChannelRejectsGlobal(t *testing.T) {
	if err := EmbedChannel(context.Background(), t.TempDir(), "global", "kid", "x"); err == nil {
		t.Fatal("expected reject")
	}
}

func openTestEnvelope(env *struct {
	V          int    `json:"v"`
	Kid        string `json:"kid"`
	EncKey     string `json:"enc_key"`
	Ciphertext string `json:"ciphertext"`
	Tag        string `json:"tag"`
}, priv *rsa.PrivateKey) ([]byte, error) {
	encKey, err := base64.StdEncoding.DecodeString(env.EncKey)
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(env.Ciphertext)
	if err != nil {
		return nil, err
	}
	tag, err := base64.StdEncoding.DecodeString(env.Tag)
	if err != nil {
		return nil, err
	}
	wrapped, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, encKey, nil)
	if err != nil {
		return nil, err
	}
	if len(wrapped) != aesKeyLen+gcmNonceLen {
		return nil, fmt.Errorf("bad wrapped")
	}
	block, err := aes.NewCipher(wrapped[:aesKeyLen])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	sealed := append(append([]byte{}, ciphertext...), tag...)
	return gcm.Open(nil, wrapped[aesKeyLen:], sealed, []byte(fmt.Sprintf("%d|%s", env.V, env.Kid)))
}
