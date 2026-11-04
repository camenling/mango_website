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
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

const (
	Rel = "website/hm-ota-channel.json"

	envelopeV   = 1
	aesKeyLen   = 32
	gcmNonceLen = 12
	gcmTagLen   = 16
)

var reChannel = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// EmbedChannel 用 kid + 公钥把渠道加密写入 app/website/hm-ota-channel.json，不签名。
func EmbedChannel(ctx context.Context, appPath, channel, kid, pubPEM string) error {
	channel = strings.ToLower(strings.TrimSpace(channel))
	kid = strings.TrimSpace(kid)
	pubPEM = strings.TrimSpace(pubPEM)
	if channel == "" || channel == "global" || !reChannel.MatchString(channel) {
		return gerror.New("渠道标识不合法")
	}
	if kid == "" || pubPEM == "" {
		return gerror.New("渠道 kid/公钥未配置")
	}
	plain, err := json.Marshal(map[string]string{"HMBuildChannel": channel})
	if err != nil {
		return err
	}
	env, err := encryptEnvelope(kid, pubPEM, plain)
	if err != nil {
		return gerror.Wrap(err, "加密渠道失败")
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return err
	}
	websiteDir := filepath.Join(appPath, "website")
	if err = gfile.Mkdir(websiteDir); err != nil {
		return gerror.Wrap(err, "创建 website 目录失败")
	}
	dst := filepath.Join(appPath, Rel)
	if err = gfile.PutBytes(dst, raw); err != nil {
		return gerror.Wrap(err, "写入 hm-ota-channel.json 失败")
	}
	g.Log().Infof(ctx, "内置渠道已写入 %s channel=%s kid=%s", Rel, channel, kid)
	return nil
}

func encryptEnvelope(kid, pubPEM string, plain []byte) (map[string]any, error) {
	pub, err := parseRSAPublicKey([]byte(pubPEM))
	if err != nil {
		return nil, err
	}
	aesKey := make([]byte, aesKeyLen)
	nonce := make([]byte, gcmNonceLen)
	if _, err = io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, err
	}
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, plain, []byte(fmt.Sprintf("%d|%s", envelopeV, strings.TrimSpace(kid))))
	if len(sealed) < gcmTagLen {
		return nil, fmt.Errorf("bad sealed")
	}
	ciphertext, tag := sealed[:len(sealed)-gcmTagLen], sealed[len(sealed)-gcmTagLen:]
	encKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, append(aesKey, nonce...), nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"v":          envelopeV,
		"kid":        kid,
		"enc_key":    base64.StdEncoding.EncodeToString(encKey),
		"ciphertext": base64.StdEncoding.EncodeToString(ciphertext),
		"tag":        base64.StdEncoding.EncodeToString(tag),
	}, nil
}

func parseRSAPublicKey(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("bad public key pem")
	}
	if pubAny, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		pub, ok := pubAny.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not rsa public key")
		}
		return pub, nil
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("bad public key")
	}
	return pub, nil
}
