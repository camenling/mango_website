package baseipa

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"mangosign/internal/pkg/siteaes"
)

var (
	downloadMu       sync.Mutex
	ErrReleaseNotFound = errors.New("not_found")
)

// Release 授权端最新已发布版本。
type Release struct {
	ReleaseID   int
	Version     string
	HmBuildDate string
	FileSHA256  string
	ClientIP    string
	FileToken   string
	Track       string
}

// Resolve 取签名底包：加密查询最新已发布版本，本地已有则用，否则走授权下载接口缓存。
func Resolve(ctx context.Context) (ipaPath string, rel *Release, err error) {
	return resolveLatest(ctx, "")
}

// LookupTrack 只查通道最新版本元数据，不下载 IPA。
func LookupTrack(ctx context.Context, track string) (*Release, error) {
	t, err := NormalizeUpdateTrack(track)
	if err != nil {
		return nil, err
	}
	rel, err := fetchLatestPublished(ctx, t)
	if err != nil {
		return nil, err
	}
	if rel == nil || rel.ReleaseID == 0 {
		return nil, nil
	}
	rel.Track = t
	rel.FileToken = ""
	return rel, nil
}

// LookupTracks 并行查询 release / beta 最新版本。
func LookupTracks(ctx context.Context) map[string]*Release {
	tracks := []string{"release", "beta"}
	out := make(map[string]*Release, len(tracks))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, track := range tracks {
		wg.Add(1)
		go func(track string) {
			defer wg.Done()
			rel, err := LookupTrack(ctx, track)
			if err != nil {
				g.Log().Warningf(ctx, "查询通道最新版本失败 track=%s: %v", track, err)
				return
			}
			if rel == nil {
				return
			}
			mu.Lock()
			out[track] = rel
			mu.Unlock()
		}(track)
	}
	wg.Wait()
	return out
}

// ResolveTrack 按更新通道取最新已发布芒果签，不回退本地 baseIpa。
func ResolveTrack(ctx context.Context, track string) (ipaPath string, rel *Release, err error) {
	return ResolveTrackBuild(ctx, track, "")
}

// ResolveTrackBuild 按通道和构建日期取已发布芒果签；build 为空则取最新。
func ResolveTrackBuild(ctx context.Context, track, build string) (ipaPath string, rel *Release, err error) {
	t, err := NormalizeUpdateTrack(track)
	if err != nil {
		return "", nil, err
	}
	build = strings.TrimSpace(build)
	if build != "" {
		rel, err = fetchPublished(ctx, t, build)
		if err != nil {
			return "", nil, err
		}
		if rel == nil || rel.ReleaseID == 0 || strings.TrimSpace(rel.HmBuildDate) != build {
			return "", nil, ErrReleaseNotFound
		}
		path, locErr := localize(ctx, rel)
		if locErr != nil {
			return "", nil, locErr
		}
		return path, rel, nil
	}
	path, rel, err := resolveLatest(ctx, t)
	if err != nil {
		return "", nil, err
	}
	if rel == nil || rel.ReleaseID == 0 {
		return "", nil, gerror.New("没有已发布的芒果签")
	}
	return path, rel, nil
}

func resolveLatest(ctx context.Context, track string) (ipaPath string, rel *Release, err error) {
	rel, err = fetchLatestPublished(ctx, track)
	if err != nil {
		if strings.TrimSpace(track) == "" {
			g.Log().Warningf(ctx, "查询最新已发布版本失败，回退 baseIpa: %v", err)
		} else {
			g.Log().Warningf(ctx, "查询通道最新版本失败 track=%s: %v", track, err)
		}
	}
	if rel != nil && rel.ReleaseID > 0 {
		path, locErr := localize(ctx, rel)
		if locErr != nil {
			return "", nil, locErr
		}
		return path, rel, nil
	}

	if strings.TrimSpace(track) != "" {
		if err != nil {
			return "", nil, err
		}
		return "", nil, gerror.New("没有已发布的芒果签")
	}

	fallback := strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.baseIpa").String())
	if fallback != "" && gfile.Exists(fallback) {
		return fallback, nil, nil
	}
	if err != nil {
		return "", nil, gerror.Wrap(err, "没有已发布版本，且本地 baseIpa 不可用")
	}
	return "", nil, gerror.New("没有已发布的签名底包，请先在授权后台发布 Mango 版本")
}

func fetchLatestPublished(ctx context.Context, track string) (*Release, error) {
	return fetchPublished(ctx, track, "")
}

func fetchPublished(ctx context.Context, track, build string) (*Release, error) {
	base := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authBaseUrl").String()), "/")
	if base == "" {
		return nil, nil
	}
	key := strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authAesKey").String())
	if key == "" {
		return nil, gerror.New("未配置 sign.authAesKey")
	}
	channel := strings.TrimSpace(track)
	if channel == "" {
		channel = strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authChannel").String())
	}
	if channel == "" {
		channel = "release"
	}
	req := map[string]string{
		"channel":   channel,
		"bundle_id": strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authBundleId").String()),
	}
	if date := strings.TrimSpace(build); date != "" {
		req["hm_build_date"] = date
	}
	reqBody, _ := json.Marshal(req)
	status, payload, err := postEncrypted(ctx, base+"/api/v1/sign/app/latest-published", key, reqBody)
	if err != nil {
		return nil, err
	}
	if status >= 400 && len(payload) == 0 {
		return nil, fmt.Errorf("授权接口 HTTP %d", status)
	}
	var envelope struct {
		Code     int             `json:"code"`
		Message  string          `json:"message"`
		ClientIP string          `json:"client_ip"`
		Data     json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, fmt.Errorf("解析授权响应失败: %w", err)
	}
	if envelope.Code != 0 {
		g.Log().Warningf(ctx, "授权接口失败 IP=%s: %s", envelope.ClientIP, envelope.Message)
		return nil, fmt.Errorf("授权接口: %s", envelope.Message)
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil, nil
	}
	var data struct {
		ReleaseID   int    `json:"release_id"`
		Version     string `json:"version"`
		HmBuildDate string `json:"hm_build_date"`
		FileSHA256  string `json:"file_sha256"`
		ClientIP    string `json:"client_ip"`
		FileToken   string `json:"file_token"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		return nil, fmt.Errorf("解析版本数据失败: %w", err)
	}
	if data.ReleaseID == 0 || strings.TrimSpace(data.FileToken) == "" {
		return nil, nil
	}
	ip := data.ClientIP
	if ip == "" {
		ip = envelope.ClientIP
	}
	g.Log().Infof(ctx, "授权识别官网 IP=%s track=%s release=%d date=%s", ip, channel, data.ReleaseID, data.HmBuildDate)
	return &Release{
		ReleaseID:   data.ReleaseID,
		Version:     data.Version,
		HmBuildDate: data.HmBuildDate,
		FileSHA256:  data.FileSHA256,
		ClientIP:    ip,
		FileToken:   data.FileToken,
		Track:       strings.TrimSpace(track),
	}, nil
}

func localize(ctx context.Context, rel *Release) (string, error) {
	if local := findExisting(ctx, rel); local != "" {
		return local, nil
	}

	downloadMu.Lock()
	defer downloadMu.Unlock()
	if local := findExisting(ctx, rel); local != "" {
		return local, nil
	}

	dest := cachePath(ctx, rel)
	if err := gfile.Mkdir(filepath.Dir(dest)); err != nil {
		return "", gerror.Wrap(err, "创建底包缓存目录失败")
	}
	tmp := dest + ".tmp"
	if err := downloadViaAPI(ctx, rel, tmp); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if expect := strings.TrimSpace(rel.FileSHA256); expect != "" {
		got, sumErr := fileSHA256(tmp)
		if sumErr != nil {
			_ = os.Remove(tmp)
			return "", sumErr
		}
		if !strings.EqualFold(got, expect) {
			_ = os.Remove(tmp)
			return "", gerror.Newf("底包 SHA256 不匹配: got %s", got)
		}
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return "", gerror.Wrap(err, "保存底包缓存失败")
	}
	g.Log().Infof(ctx, "已从授权接口缓存底包 %d_%s -> %s", rel.ReleaseID, rel.HmBuildDate, dest)
	return dest, nil
}

func downloadViaAPI(ctx context.Context, rel *Release, dest string) error {
	base := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authBaseUrl").String()), "/")
	key := strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authAesKey").String())
	reqBody, _ := json.Marshal(map[string]interface{}{
		"release_id": rel.ReleaseID,
		"token":      rel.FileToken,
	})
	cipher, err := siteaes.Encrypt(key, string(reqBody))
	if err != nil {
		return err
	}
	wrap, _ := json.Marshal(map[string]string{"cipher": cipher})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v1/sign/app/latest-published/file", bytes.NewReader(wrap))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 15 * time.Minute}).Do(req)
	if err != nil {
		return gerror.Wrap(err, "下载底包失败")
	}
	defer resp.Body.Close()
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "json") || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		plain, decErr := decodeAuthBody(key, body)
		if decErr == nil {
			var envelope struct {
				Message  string `json:"message"`
				ClientIP string `json:"client_ip"`
			}
			_ = json.Unmarshal(plain, &envelope)
			if envelope.Message != "" {
				g.Log().Warningf(ctx, "下载底包失败 IP=%s: %s", envelope.ClientIP, envelope.Message)
				return gerror.New(envelope.Message)
			}
		}
		return gerror.Newf("下载底包失败: HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return gerror.Wrap(err, "写入底包失败")
	}
	if ip := strings.TrimSpace(resp.Header.Get("X-Mango-Client-IP")); ip != "" {
		rel.ClientIP = ip
	}
	return nil
}

func postEncrypted(ctx context.Context, api, key string, plain []byte) (status int, payload []byte, err error) {
	cipher, err := siteaes.Encrypt(key, string(plain))
	if err != nil {
		return 0, nil, err
	}
	wrap, _ := json.Marshal(map[string]string{"cipher": cipher})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api, bytes.NewReader(wrap))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	plainResp, decErr := decodeAuthBody(key, body)
	if decErr != nil {
		return resp.StatusCode, nil, decErr
	}
	return resp.StatusCode, plainResp, nil
}

func decodeAuthBody(key string, body []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("授权响应为空")
	}
	var wrap struct {
		Cipher string `json:"cipher"`
	}
	if json.Unmarshal(trimmed, &wrap) == nil && strings.TrimSpace(wrap.Cipher) != "" {
		plain, err := siteaes.Decrypt(key, wrap.Cipher)
		if err != nil {
			return nil, fmt.Errorf("授权响应解密失败: %w", err)
		}
		return []byte(plain), nil
	}
	return nil, fmt.Errorf("授权响应未加密")
}

func findExisting(ctx context.Context, rel *Release) string {
	cache := cachePath(ctx, rel)
	if fileUsable(cache) && sha256OK(cache, rel.FileSHA256) {
		return cache
	}
	return ""
}

func cachePath(ctx context.Context, rel *Release) string {
	dir := strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.ipaCacheDir").String())
	if dir == "" {
		dir = "resource/base/cache"
	}
	name := fmt.Sprintf("%d_%s.ipa", rel.ReleaseID, sanitizeName(rel.HmBuildDate))
	if track := sanitizeName(rel.Track); track != "" && track != "unknown" {
		name = fmt.Sprintf("%s_%s", track, name)
	}
	return filepath.Join(dir, name)
}

// NormalizeUpdateTrack 更新通道：release / beta，空值按 release。
func NormalizeUpdateTrack(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "release":
		return "release", nil
	case "beta":
		return "beta", nil
	default:
		return "", gerror.New("bad track")
	}
}

func uploadRelPath(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	path := rawURL
	if u, err := url.Parse(rawURL); err == nil && u.Path != "" && (u.Scheme == "http" || u.Scheme == "https") {
		path = u.Path
	}
	path = strings.TrimPrefix(path, "/")
	idx := strings.Index(strings.ToLower(path), "upload_file/")
	if idx < 0 {
		return ""
	}
	return path[idx:]
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func sha256OK(path, expect string) bool {
	expect = strings.TrimSpace(expect)
	if expect == "" {
		return true
	}
	got, err := fileSHA256(path)
	if err != nil {
		return false
	}
	return strings.EqualFold(got, expect)
}

func fileUsable(path string) bool {
	if path == "" || !gfile.Exists(path) {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}

func sanitizeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}
