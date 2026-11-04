package controller

import (
	"html"

	"github.com/gogf/gf/v2/net/ghttp"

	installpkg "mangosign/internal/pkg/install"
)

var InstallShort = cInstallShort{}

type cInstallShort struct{}

// ShortInstallItms GET /i/:key → 302 到 itms-services（扫码/短链安装）
func (c *cInstallShort) ShortInstallItms(r *ghttp.Request) {
	key := r.Get("key").String()
	plistURL, err := installpkg.GetShortInstallPlist(r.Context(), key)
	if err != nil {
		r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		r.Response.Write(installErrorHTML(err.Error()))
		return
	}
	target := installpkg.ItmsInstallURL(plistURL)
	r.Response.RedirectTo(target)
}

// GetInstallPlist GET /getInstallPlist/:uuid → OTA manifest
func (c *cInstallShort) GetInstallPlist(r *ghttp.Request) {
	uuid := r.Get("uuid").String()
	content, err := installpkg.LoadPlistInfo(r.Context(), uuid)
	if err != nil {
		r.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		r.Response.Write(installErrorHTML(err.Error()))
		return
	}
	r.Response.Header().Set("Content-Type", "application/xml; charset=utf-8")
	r.Response.Write(content)
}

func installErrorHTML(msg string) []byte {
	return []byte(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>安装失败</title></head><body style="font-family:-apple-system,sans-serif;padding:24px;text-align:center"><p>` + html.EscapeString(msg) + `</p></body></html>`)
}
