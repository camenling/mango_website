package controller

import (
	"encoding/base64"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"mangosign/internal/pkg/tabicon"
)

func TabIconsCatalog(r *ghttp.Request) {
	ctx := r.Context()
	body := r.GetBody()
	if len(body) == 0 {
		r.Response.WriteStatus(400)
		return
	}
	opened, err := tabicon.OpenEnvelope(ctx, json.RawMessage(body))
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标拆信封失败: %v", err)
		r.Response.WriteStatus(403, "无法确认这台设备")
		return
	}
	payload, err := tabicon.Payload(ctx, opened.Revision, opened.Page, opened.PageSize)
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标组包失败: %v", err)
		r.Response.WriteStatus(503, err.Error())
		return
	}
	pack, err := tabicon.PackPreview(payload)
	if err != nil {
		r.Response.WriteStatus(503, err.Error())
		return
	}
	sealed, err := tabicon.SealReply(ctx, json.RawMessage(body), pack.Meta)
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标密封失败: %v", err)
		r.Response.WriteStatus(503, err.Error())
		return
	}
	sealed["preview"] = base64.StdEncoding.EncodeToString(pack.File)
	r.Response.WriteJsonExit(sealed)
}

func TabIconsFile(r *ghttp.Request) {
	ctx := r.Context()
	body := r.GetBody()
	if len(body) == 0 {
		r.Response.WriteStatus(400)
		return
	}
	opened, err := tabicon.OpenEnvelope(ctx, json.RawMessage(body))
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标拆信封失败: %v", err)
		r.Response.WriteStatus(403, "无法确认这台设备")
		return
	}
	pack, err := tabicon.PackFile(ctx, opened.ThemeKey, opened.Revision)
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标组包失败: %v", err)
		r.Response.WriteStatus(503, err.Error())
		return
	}
	sealed, err := tabicon.SealReply(ctx, json.RawMessage(body), pack.Meta)
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标密封失败: %v", err)
		r.Response.WriteStatus(503, err.Error())
		return
	}
	sealed["file"] = base64.StdEncoding.EncodeToString(pack.File)
	r.Response.WriteJsonExit(sealed)
}

func replyTabIcons(r *ghttp.Request, build func(opened tabicon.Opened) (map[string]any, error)) {
	ctx := r.Context()
	body := r.GetBody()
	if len(body) == 0 {
		r.Response.WriteStatus(400)
		return
	}
	opened, err := tabicon.OpenEnvelope(ctx, json.RawMessage(body))
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标拆信封失败: %v", err)
		r.Response.WriteStatus(403, "无法确认这台设备")
		return
	}
	payload, err := build(opened)
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标组包失败: %v", err)
		r.Response.WriteStatus(503, err.Error())
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		r.Response.WriteStatus(503, "图标列表暂不可用")
		return
	}
	sealed, err := tabicon.SealReply(ctx, json.RawMessage(body), raw)
	if err != nil {
		g.Log().Warningf(ctx, "底栏图标密封失败: %v", err)
		r.Response.WriteStatus(503, err.Error())
		return
	}
	r.Response.WriteJsonExit(sealed)
}
