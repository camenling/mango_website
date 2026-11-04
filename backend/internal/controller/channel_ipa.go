package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/guid"

	"mangosign/internal/pkg/apptheme"
	"mangosign/internal/pkg/baseipa"
	"mangosign/internal/pkg/ipazip"
	"mangosign/internal/pkg/otachannel"
)

// channelIpaTheme 下载请求体里可选的 theme 字段（不加密，和信封字段并列）。
// 二选一：style/accent/glass 用官网预设；package 为主题配置页导出的 .hmtheme Base64。
type channelIpaTheme struct {
	Style   int    `json:"style"`
	Accent  string `json:"accent"`
	Glass   *bool  `json:"glass"`
	Package string `json:"package"`
}

// splitChannelIpaTheme 把 theme 从请求体里摘出来，剩余字段原样转发给鉴权服务。
func splitChannelIpaTheme(body []byte) ([]byte, *channelIpaTheme, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, nil, err
	}
	themeRaw, ok := raw["theme"]
	if !ok || len(bytes.TrimSpace(themeRaw)) == 0 || string(bytes.TrimSpace(themeRaw)) == "null" {
		return body, nil, nil
	}
	var theme channelIpaTheme
	if err := json.Unmarshal(themeRaw, &theme); err != nil {
		return nil, nil, err
	}
	delete(raw, "theme")
	envelope, err := json.Marshal(raw)
	if err != nil {
		return nil, nil, err
	}
	return envelope, &theme, nil
}

// applyChannelIpaTheme 按 theme 字段把主题写进 .app；校验失败返回 bad_theme。
func applyChannelIpaTheme(ctx context.Context, appPath string, theme *channelIpaTheme) error {
	if theme == nil {
		return nil
	}
	if pkg := strings.TrimSpace(theme.Package); pkg != "" {
		if i := strings.IndexByte(pkg, ','); i >= 0 && strings.HasPrefix(pkg, "data:") {
			pkg = pkg[i+1:]
		}
		if len(pkg) > apptheme.MaxPackageBytes*4/3+4 {
			return apptheme.ErrBadPackage
		}
		data, err := base64.StdEncoding.DecodeString(pkg)
		if err != nil {
			return apptheme.ErrBadPackage
		}
		return apptheme.EmbedPackage(ctx, appPath, data)
	}
	if theme.Style < apptheme.MinID || theme.Style > apptheme.MaxID {
		return apptheme.ErrBadPackage
	}
	glass := true
	if theme.Glass != nil {
		glass = *theme.Glass
	}
	_, err := apptheme.ApplyTheme(ctx, appPath, apptheme.Config{
		Theme:  theme.Style,
		Accent: theme.Accent,
		Glass:  glass,
	})
	return err
}

type channelIpaAuthReply struct {
	OK        bool   `json:"ok"`
	Channel   string `json:"channel"`
	Kid       string `json:"kid"`
	PublicKey string `json:"public_key"`
	Track     string `json:"track"`
	Build     string `json:"build"`
	Error     string `json:"error"`
}

// ChannelIpa 校验 kid+密钥信封后，按 track+build 取对应芒果签，只内置渠道、不签名，再打包返回。
func ChannelIpa(r *ghttp.Request) {
	ctx := r.Context()
	body := r.GetBody()
	if len(bytes.TrimSpace(body)) == 0 {
		writeEncryptCertError(r, http.StatusBadRequest, "bad_request")
		return
	}
	envelope, theme, err := splitChannelIpaTheme(body)
	if err != nil {
		writeEncryptCertError(r, http.StatusBadRequest, "bad_request")
		return
	}
	reply, status, err := verifyChannelIpa(ctx, envelope)
	if err != nil {
		writeChannelIpaAuthError(r, status, err)
		return
	}

	ipaPath, rel, err := baseipa.ResolveTrackBuild(ctx, reply.Track, reply.Build)
	if err != nil || ipaPath == "" {
		if errors.Is(err, baseipa.ErrReleaseNotFound) {
			writeEncryptCertError(r, http.StatusNotFound, "not_found")
			return
		}
		g.Log().Warning(ctx, "channel ipa resolve:", err)
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}

	work, extract, outIpa, err := channelIpaWorkDir(ctx)
	if err != nil {
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	defer os.RemoveAll(work)

	if err = ipazip.Unzip(ipaPath, extract); err != nil {
		g.Log().Warning(ctx, "channel ipa unzip:", err)
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	appPath, err := ipazip.FindApp(extract)
	if err != nil {
		g.Log().Warning(ctx, "channel ipa app:", err)
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	if err = otachannel.EmbedChannel(ctx, appPath, reply.Channel, reply.Kid, reply.PublicKey); err != nil {
		g.Log().Warning(ctx, "channel ipa embed:", err)
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	if err = applyChannelIpaTheme(ctx, appPath, theme); err != nil {
		if errors.Is(err, apptheme.ErrBadPackage) {
			writeEncryptCertError(r, http.StatusBadRequest, "bad_theme")
			return
		}
		g.Log().Warning(ctx, "channel ipa theme:", err)
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	if err = ipazip.ZipDir(extract, outIpa); err != nil {
		g.Log().Warning(ctx, "channel ipa zip:", err)
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}

	build := ""
	if rel != nil {
		build = rel.HmBuildDate
	}
	name := channelIpaFileName(reply.Channel, reply.Track, build)
	r.Response.Header().Set("Content-Type", "application/octet-stream")
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	r.Response.ServeFile(outIpa)
	r.ExitAll()
}

// ChannelIpaCheck 与下载同一套 kid 信封，验过后再比对当前 build 是否最新。
func ChannelIpaCheck(r *ghttp.Request) {
	ctx := r.Context()
	body := r.GetBody()
	if len(bytes.TrimSpace(body)) == 0 {
		writeEncryptCertError(r, http.StatusBadRequest, "bad_request")
		return
	}
	reply, status, err := verifyChannelIpa(ctx, body)
	if err != nil {
		writeChannelIpaAuthError(r, status, err)
		return
	}
	rel, err := baseipa.LookupTrack(ctx, reply.Track)
	if err != nil {
		g.Log().Warning(ctx, "channel ipa check:", err)
		writeEncryptCertError(r, http.StatusServiceUnavailable, "unavailable")
		return
	}
	var latest g.Map
	isLatest := true
	if rel != nil && rel.ReleaseID > 0 {
		date := strings.TrimSpace(rel.HmBuildDate)
		latest = g.Map{
			"track":         rel.Track,
			"version":       strings.TrimSpace(rel.Version),
			"hm_build_date": date,
		}
		if date != "" {
			isLatest = !(date > reply.Build)
		}
	}
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.WriteJsonExit(g.Map{
		"ok":        true,
		"track":     reply.Track,
		"channel":   reply.Channel,
		"build":     reply.Build,
		"is_latest": isLatest,
		"latest":    latest,
	})
}

func writeChannelIpaAuthError(r *ghttp.Request, status int, err error) {
	msg := err.Error()
	switch msg {
	case "bad_request", "decrypt_failed", "bad_payload", "denied", "not_found":
	default:
		msg = "unavailable"
	}
	if status == 0 {
		status = http.StatusServiceUnavailable
	}
	writeEncryptCertError(r, status, msg)
}

func verifyChannelIpa(ctx context.Context, body []byte) (*channelIpaAuthReply, int, error) {
	base := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authBaseUrl").String()), "/")
	if base == "" {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("unavailable")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v1/device/channel-ipa", bytes.NewReader(body))
	if err != nil {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("unavailable")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		g.Log().Warning(ctx, "channel ipa auth:", err)
		return nil, http.StatusServiceUnavailable, fmt.Errorf("unavailable")
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("unavailable")
	}
	var reply channelIpaAuthReply
	if err = json.Unmarshal(out, &reply); err != nil {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("unavailable")
	}
	if resp.StatusCode != http.StatusOK || !reply.OK {
		msg := strings.TrimSpace(reply.Error)
		if msg == "" {
			msg = "unavailable"
		}
		status := resp.StatusCode
		if status == 0 {
			status = http.StatusServiceUnavailable
		}
		return nil, status, fmt.Errorf("%s", msg)
	}
	if strings.TrimSpace(reply.Channel) == "" || strings.TrimSpace(reply.Kid) == "" || strings.TrimSpace(reply.PublicKey) == "" {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("unavailable")
	}
	track, err := baseipa.NormalizeUpdateTrack(reply.Track)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("bad_payload")
	}
	build := strings.TrimSpace(reply.Build)
	if build == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("bad_payload")
	}
	reply.Track = track
	reply.Build = build
	return &reply, http.StatusOK, nil
}

func channelIpaWorkDir(ctx context.Context) (work, extract, outIpa string, err error) {
	root := strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.tempDir").String())
	if root == "" {
		root = "resource/temp"
	}
	work = filepath.Join(root, "channel-ipa", guid.S())
	extract = filepath.Join(work, "extract")
	outIpa = filepath.Join(work, "out.ipa")
	if err = gfile.Mkdir(extract); err != nil {
		return "", "", "", err
	}
	return work, extract, outIpa, nil
}

func channelIpaFileName(channel, track, build string) string {
	ch := baseipaSanitize(channel)
	tr := baseipaSanitize(track)
	bd := baseipaSanitize(build)
	if bd == "" || bd == "unknown" {
		bd = "latest"
	}
	return fmt.Sprintf("mango-%s-%s-%s.ipa", ch, tr, bd)
}

func baseipaSanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}
