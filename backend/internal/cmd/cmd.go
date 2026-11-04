package cmd

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gfile"

	"mangosign/internal/controller"
	"mangosign/internal/pkg/tabicon"
	"mangosign/internal/service"
)

var Main = gcmd.Command{
	Name:  "main",
	Usage: "main",
	Brief: "芒果签 iOS Signing Server",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		cfg := g.Cfg()
		cacheDir := cfg.MustGet(ctx, "sign.ipaCacheDir").String()
		if cacheDir == "" {
			cacheDir = "resource/base/cache"
		}
		for _, dir := range []string{
			cfg.MustGet(ctx, "sign.tempDir").String(),
			cfg.MustGet(ctx, "sign.outputDir").String(),
			cfg.MustGet(ctx, "sign.signTempDir").String(),
			"resource/base",
			cacheDir,
			"resource/public",
		} {
			if !gfile.Exists(dir) {
				if err = gfile.Mkdir(dir); err != nil {
					return err
				}
			}
		}

		s := g.Server()
		s.SetReadTimeout(2 * time.Minute)
		s.SetWriteTimeout(20 * time.Minute)
		s.SetIdleTimeout(2 * time.Minute)
		s.Use(ghttp.MiddlewareCORS)

		s.Group("/api", func(group *ghttp.RouterGroup) {
			group.Middleware(ghttp.MiddlewareHandlerResponse)
			group.Bind(
				controller.Sign,
			)
		})

		// 加密证书：原样透传到授权，不走 JSON 包装
		s.Group("/api", func(group *ghttp.RouterGroup) {
			group.POST("/sign/encrypt-cert", controller.EncryptCertProxy)
			group.POST("/sign/channel-blacklist", controller.ChannelBlacklistProxy)
			group.POST("/sign/channel-unblacklist", controller.ChannelUnblacklistProxy)
			group.POST("/sign/channel-ipa", controller.ChannelIpa)
			group.POST("/sign/channel-ipa-check", controller.ChannelIpaCheck)
			group.POST("/sign/blacklist-lookup", controller.BlacklistLookupProxy)
		})

		s.Group("/api", func(group *ghttp.RouterGroup) {
			group.POST("/tab-icons/catalog", controller.TabIconsCatalog)
			group.POST("/tab-icons/file", controller.TabIconsFile)
		})

		// 公众号凭证：微信网页授权 + 代用户调授权系统（IP + AES）
		s.Group("/api/wx-cert", func(group *ghttp.RouterGroup) {
			group.GET("/login", controller.WxCertLogin)
			group.GET("/callback", controller.WxCertCallback)
			group.POST("/{action}", controller.WxCertAction)
		})

		// 安装短链与 OTA plist（不走 JSON 包装，对齐 sign/backend）
		s.BindHandler("/i/{key}", controller.InstallShort.ShortInstallItms)
		s.BindHandler("/getInstallPlist/{uuid}", controller.InstallShort.GetInstallPlist)

		// SPA 页面回退（静态资源由 serverRoot 自动提供）。
		// 前端 router/index.js 新增页面时这里要同步加，否则直接访问或刷新该地址会 404。
		for _, p := range []string{"/", "/install", "/theme", "/docs", "/lookup", "/channel-ipa", "/wx-cert"} {
			path := p
			s.BindHandler(path, func(r *ghttp.Request) {
				r.Response.ServeFile("resource/public/index.html")
			})
		}

		service.Sign().StartCleanupWorker(ctx)
		go tabicon.Warm(ctx)
		s.Run()
		return nil
	},
}
