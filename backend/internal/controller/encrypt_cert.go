package controller

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const certEncryptClientIPHdr = "X-HM-Client-IP"

// EncryptCertProxy 把客户加密信封原样转到授权 /api/v1/device/encrypt-cert，并带上访问者 IP。
func EncryptCertProxy(r *ghttp.Request) {
	proxyAuthDevice(r, "encrypt-cert")
}

// ChannelBlacklistProxy 把客户加密信封原样转到授权 /api/v1/device/channel-blacklist，并带上访问者 IP。
func ChannelBlacklistProxy(r *ghttp.Request) {
	proxyAuthDevice(r, "channel-blacklist")
}

// ChannelUnblacklistProxy 把客户加密信封原样转到授权 /api/v1/device/channel-unblacklist，并带上访问者 IP。
func ChannelUnblacklistProxy(r *ghttp.Request) {
	proxyAuthDevice(r, "channel-unblacklist")
}

// BlacklistLookupProxy 查询 UDID 拉黑情况，转到授权 /api/v1/device/blacklist-lookup。
func BlacklistLookupProxy(r *ghttp.Request) {
	proxyAuthDevice(r, "blacklist-lookup")
}

func proxyAuthDevice(r *ghttp.Request, path string) {
	ctx := r.Context()
	base := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authBaseUrl").String()), "/")
	if base == "" {
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	body := r.GetBody()
	if len(bytes.TrimSpace(body)) == 0 {
		writeEncryptCertError(r, http.StatusBadRequest, "bad_request")
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v1/device/"+path, bytes.NewReader(body))
	if err != nil {
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(certEncryptClientIPHdr, r.GetClientIp())
	resp, err := (&http.Client{Timeout: 90 * time.Second}).Do(req)
	if err != nil {
		g.Log().Warning(ctx, path+" proxy:", err)
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	status := resp.StatusCode
	if status == 0 {
		status = http.StatusServiceUnavailable
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/json"
	}
	r.Response.Header().Set("Content-Type", ct)
	r.Response.Status = status
	r.Response.Write(out)
	r.ExitAll()
}

func writeEncryptCertError(r *ghttp.Request, status int, msg string) {
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.Status = status
	r.Response.WriteJsonExit(g.Map{"error": msg})
}
