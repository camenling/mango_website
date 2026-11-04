// Package wxcert 调授权系统的公众号凭证接口（/api/v1/sign/app/wx-cert/*）。
// 和底包、底栏图标一样：请求体 AES 加密，授权侧校验本机出口 IP 白名单，响应也是密文。
package wxcert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"mangosign/internal/pkg/siteaes"
)

// Reply 授权返回的明文。Code 0 成功，401 表示 ticket 失效需要重新微信授权。
type Reply struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

const CodeTicketExpired = 401

func Call(ctx context.Context, action string, payload any) (*Reply, error) {
	base := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authBaseUrl").String()), "/")
	key := strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authAesKey").String())
	if base == "" || key == "" {
		return nil, fmt.Errorf("未配置授权地址")
	}
	plain, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	cipher, err := siteaes.Encrypt(key, string(plain))
	if err != nil {
		return nil, err
	}
	wrap, _ := json.Marshal(map[string]string{"cipher": cipher})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v1/sign/app/wx-cert/"+action, bytes.NewReader(wrap))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var env struct {
		Cipher string `json:"cipher"`
	}
	if json.Unmarshal(bytes.TrimSpace(body), &env) != nil || strings.TrimSpace(env.Cipher) == "" {
		return nil, fmt.Errorf("授权响应未加密: HTTP %d", resp.StatusCode)
	}
	text, err := siteaes.Decrypt(key, env.Cipher)
	if err != nil {
		return nil, fmt.Errorf("授权响应解密失败: %w", err)
	}
	var out Reply
	if err = json.Unmarshal([]byte(text), &out); err != nil {
		return nil, fmt.Errorf("解析授权响应失败: %w", err)
	}
	return &out, nil
}
