package controller

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"mangosign/internal/pkg/wxcert"
)

// 公众号凭证页 /wx-cert 的后端：微信网页授权换 ticket，之后代用户调授权系统。
// ticket 只放在 HttpOnly Cookie，前端拿不到；openid 不出授权系统。

const (
	wxCertTicketCookie = "hm_wx_cert"
	wxCertStateCookie  = "hm_wx_state"
	wxCertCookiePath   = "/api/wx-cert"
	wxCertPage         = "/wx-cert"
)

var wxCertActions = map[string]bool{
	"list": true, "create": true, "claim": true, "channel": true,
	"ip-add": true, "ip-remove": true, "reset": true, "logout": true,
}

// WxCertLogin GET /api/wx-cert/login：跳去微信网页授权。
func WxCertLogin(r *ghttp.Request) {
	ctx := r.Context()
	reply, err := wxcert.Call(ctx, "oauth-url", g.Map{})
	if err != nil || reply.Code != 0 {
		g.Log().Warningf(ctx, "wx-cert oauth-url: %v %+v", err, reply)
		wxCertBackToPage(r, "unavailable")
		return
	}
	var data struct {
		URL   string `json:"url"`
		State string `json:"state"`
	}
	if json.Unmarshal(reply.Data, &data) != nil || data.URL == "" || data.State == "" {
		wxCertBackToPage(r, "unavailable")
		return
	}
	wxCertSetCookie(r, wxCertStateCookie, data.State, 300)
	r.Response.RedirectTo(data.URL, http.StatusFound)
}

// WxCertCallback GET /api/wx-cert/callback：微信带 code 回来，换成 ticket。
func WxCertCallback(r *ghttp.Request) {
	ctx := r.Context()
	code := strings.TrimSpace(r.Get("code").String())
	state := strings.TrimSpace(r.Get("state").String())
	cookieState := strings.TrimSpace(r.Cookie.Get(wxCertStateCookie).String())
	wxCertSetCookie(r, wxCertStateCookie, "", -1)
	if code == "" || state == "" || state != cookieState {
		wxCertBackToPage(r, "auth")
		return
	}
	reply, err := wxcert.Call(ctx, "oauth-exchange", g.Map{"code": code, "state": state})
	if err != nil || reply.Code != 0 {
		g.Log().Warningf(ctx, "wx-cert oauth-exchange: %v %+v", err, reply)
		wxCertBackToPage(r, "auth")
		return
	}
	var data struct {
		Ticket    string `json:"ticket"`
		ExpiresIn int    `json:"expires_in"`
	}
	if json.Unmarshal(reply.Data, &data) != nil || data.Ticket == "" {
		wxCertBackToPage(r, "auth")
		return
	}
	wxCertSetCookie(r, wxCertTicketCookie, data.Ticket, data.ExpiresIn)
	r.Response.RedirectTo(wxCertPage, http.StatusFound)
}

// WxCertAction POST /api/wx-cert/{action}：带上 Cookie 里的 ticket 转给授权系统。
func WxCertAction(r *ghttp.Request) {
	ctx := r.Context()
	r.Response.Header().Set("Cache-Control", "no-store")
	action := r.Get("action").String()
	if !wxCertActions[action] {
		r.Response.WriteStatusExit(http.StatusNotFound)
		return
	}
	ticket := strings.TrimSpace(r.Cookie.Get(wxCertTicketCookie).String())
	if ticket == "" {
		r.Response.WriteJsonExit(g.Map{"code": wxcert.CodeTicketExpired, "message": "请先在微信里授权"})
		return
	}
	payload := map[string]any{}
	if body := r.GetBody(); len(strings.TrimSpace(string(body))) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			r.Response.WriteJsonExit(g.Map{"code": 1, "message": "请求格式不正确"})
			return
		}
	}
	payload["ticket"] = ticket
	reply, err := wxcert.Call(ctx, action, payload)
	if err != nil {
		g.Log().Warningf(ctx, "wx-cert %s: %v", action, err)
		r.Response.WriteJsonExit(g.Map{"code": 1, "message": "服务暂时不可用，请稍后再试"})
		return
	}
	if action == "logout" || reply.Code == wxcert.CodeTicketExpired {
		wxCertSetCookie(r, wxCertTicketCookie, "", -1)
	}
	r.Response.WriteJsonExit(reply)
}

func wxCertBackToPage(r *ghttp.Request, reason string) {
	r.Response.RedirectTo(wxCertPage+"?error="+url.QueryEscape(reason), http.StatusFound)
}

func wxCertSetCookie(r *ghttp.Request, name, value string, maxAge int) {
	http.SetCookie(r.Response.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     wxCertCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"),
		SameSite: http.SameSiteLaxMode,
	})
}
