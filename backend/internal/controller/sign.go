package controller

import (
	"context"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"mangosign/internal/model"
	"mangosign/internal/pkg/baseipa"
	"mangosign/internal/service"
)

var Sign = cSign{}

type cSign struct{}

type SignReq struct {
	g.Meta `path:"/sign/sign" method:"post" mime:"multipart/form-data,application/json" tags:"Sign" summary:"上传证书并签名安装"`
	CertBase64Payload
	Password  string `json:"password" form:"password" v:"required|length:1,128#请输入证书密码|密码长度不合法"`
	Theme     int    `json:"theme" form:"theme" d:"1"`
	Accent    string `json:"accent" form:"accent"`
	Glass     *bool  `json:"glass" form:"glass"`
	AppSource string `json:"appSource" form:"appSource"`
}

type CertCheckReq struct {
	g.Meta `path:"/sign/check-cert" method:"post" mime:"multipart/form-data,application/json" tags:"Sign" summary:"查询证书状态"`
	CertBase64Payload
	Password string `json:"password" form:"password" v:"required|length:1,128#请输入证书密码|密码长度不合法"`
}

type SignRes struct {
	*model.SignResult
}

type CertCheckRes struct {
	*model.CertInfo
}

type HealthRes struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

type HealthReq struct {
	g.Meta `path:"/sign/health" method:"get" tags:"Sign" summary:"健康检查"`
}

type ChannelIpaTracksReq struct {
	g.Meta `path:"/sign/channel-ipa-tracks" method:"get" tags:"Sign" summary:"当前 release/beta 最新版本"`
}

type ChannelIpaTrackItem struct {
	Track       string `json:"track"`
	Version     string `json:"version"`
	HmBuildDate string `json:"hm_build_date"`
}

type ChannelIpaTracksRes struct {
	Release *ChannelIpaTrackItem `json:"release"`
	Beta    *ChannelIpaTrackItem `json:"beta"`
}

type ParseZipReq struct {
	g.Meta `path:"/sign/parse-zip" method:"post" mime:"multipart/form-data" tags:"Sign" summary:"解析 ZIP 压缩包中的证书文件"`
}

type ParseZipRes struct {
	P12Name       string `json:"p12Name"`
	ProvisionName string `json:"provisionName"`
	Message       string `json:"message"`
}

func (c *cSign) Sign(ctx context.Context, req *SignReq) (res *SignRes, err error) {
	r := g.RequestFromCtx(ctx)

	password := strings.TrimSpace(req.Password)
	if password == "" {
		password = strings.TrimSpace(r.Get("password").String())
	}
	if err = validatePassword(password); err != nil {
		return nil, err
	}

	taskId := guid.S()
	tempDir := g.Cfg().MustGet(ctx, "sign.tempDir").String()
	taskDir := gfile.Join(tempDir, taskId)
	if err = gfile.Mkdir(taskDir); err != nil {
		return nil, gerror.Wrap(err, "创建临时目录失败")
	}

	certs, err := resolveCertUploads(r, taskDir, req.CertBase64Payload)
	if err != nil {
		return nil, err
	}

	theme := req.Theme
	if theme <= 0 {
		theme = r.Get("theme").Int()
	}
	if theme <= 0 {
		theme = 1
	}

	accent := strings.TrimSpace(req.Accent)
	if accent == "" {
		accent = strings.TrimSpace(r.Get("accent").String())
	}
	glass := true
	if req.Glass != nil {
		glass = *req.Glass
	} else if v := r.Get("glass"); !v.IsNil() {
		glass = v.Bool()
	}

	appSource := strings.TrimSpace(req.AppSource)
	if appSource == "" {
		appSource = strings.TrimSpace(r.Get("appSource").String())
	}

	result, err := service.Sign().SignAndInstall(ctx, &model.SignInput{
		P12Path:             certs.p12Path,
		MobileprovisionPath: certs.provisionPath,
		Password:            password,
		Theme:               theme,
		Accent:              accent,
		Glass:               glass,
		AppSource:           appSource,
	})
	if err != nil {
		return nil, err
	}
	result.P12Name = certs.p12Name
	result.ProvisionName = certs.provisionName
	if certs.fromZip {
		result.Message = "已从 ZIP 识别 p12 与描述文件并完成签名"
	}
	return &SignRes{result}, nil
}

func (c *cSign) CheckCert(ctx context.Context, req *CertCheckReq) (res *CertCheckRes, err error) {
	r := g.RequestFromCtx(ctx)

	password := strings.TrimSpace(req.Password)
	if password == "" {
		password = strings.TrimSpace(r.Get("password").String())
	}
	if err = validatePassword(password); err != nil {
		return nil, err
	}

	taskId := guid.S()
	tempDir := g.Cfg().MustGet(ctx, "sign.tempDir").String()
	taskDir := gfile.Join(tempDir, "check_"+taskId)
	if err = gfile.Mkdir(taskDir); err != nil {
		return nil, gerror.Wrap(err, "创建临时目录失败")
	}
	defer gfile.Remove(taskDir)

	certs, err := resolveCertUploads(r, taskDir, req.CertBase64Payload)
	if err != nil {
		return nil, err
	}

	info, err := service.Sign().CheckCert(ctx, &model.CertCheckInput{
		P12Path:             certs.p12Path,
		MobileprovisionPath: certs.provisionPath,
		Password:            password,
	})
	if err != nil {
		return nil, err
	}
	return &CertCheckRes{info}, nil
}

func (c *cSign) Health(ctx context.Context, req *HealthReq) (res *HealthRes, err error) {
	return &HealthRes{
		Status: "ok",
		Time:   gtime.Now().String(),
	}, nil
}

func (c *cSign) ChannelIpaTracks(ctx context.Context, _ *ChannelIpaTracksReq) (*ChannelIpaTracksRes, error) {
	found := baseipa.LookupTracks(ctx)
	return &ChannelIpaTracksRes{
		Release: channelIpaTrackItem(found["release"]),
		Beta:    channelIpaTrackItem(found["beta"]),
	}, nil
}

func channelIpaTrackItem(rel *baseipa.Release) *ChannelIpaTrackItem {
	if rel == nil || rel.ReleaseID == 0 {
		return nil
	}
	return &ChannelIpaTrackItem{
		Track:       rel.Track,
		Version:     strings.TrimSpace(rel.Version),
		HmBuildDate: strings.TrimSpace(rel.HmBuildDate),
	}
}

func (c *cSign) ParseZip(ctx context.Context, req *ParseZipReq) (res *ParseZipRes, err error) {
	r := g.RequestFromCtx(ctx)
	zipFile := r.GetUploadFile("certZip")
	if zipFile == nil {
		return nil, gerror.New("请上传 ZIP 压缩包")
	}
	if !isZipFile(zipFile.Filename) {
		return nil, gerror.New("请上传 .zip 格式压缩包")
	}

	taskId := guid.S()
	tempDir := g.Cfg().MustGet(ctx, "sign.tempDir").String()
	taskDir := gfile.Join(tempDir, "parse_"+taskId)
	defer gfile.Remove(taskDir)

	certs, err := resolveCertUploads(r, taskDir, CertBase64Payload{})
	if err != nil {
		return nil, err
	}

	return &ParseZipRes{
		P12Name:       certs.p12Name,
		ProvisionName: certs.provisionName,
		Message:       "已识别压缩包内的证书文件",
	}, nil
}

func validatePassword(password string) error {
	if password == "" {
		return gerror.New("请输入证书密码")
	}
	// 仅允许字母与数字，防止命令注入
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, password); !matched {
		return gerror.New("证书密码仅允许字母与数字")
	}
	return nil
}

func saveUpload(file *ghttp.UploadFile, path string) error {
	_, err := file.Save(path, true)
	return err
}

// Upload 兼容 multipart 直接上传（部分客户端使用）
func (c *cSign) Upload(r *ghttp.Request) {
	var req SignReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 1, "message": err.Error()})
	}
	res, err := c.Sign(r.Context(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 1, "message": err.Error()})
	}
	r.Response.WriteJsonExit(g.Map{"code": 0, "data": res})
}
