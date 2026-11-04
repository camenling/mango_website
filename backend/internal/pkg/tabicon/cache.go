package tabicon

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	"mangosign/internal/pkg/siteaes"
)

const refreshInterval = 10 * time.Minute

var slots = []string{"app", "sign", "cert", "file", "setting"}

type Theme struct {
	ID       int    `json:"id"`
	ThemeKey string `json:"theme_key"`
	Name     string `json:"name"`
	Summary  string `json:"summary"`
	Sort     int    `json:"sort"`
	Revision int    `json:"revision"`
	SHA256   string `json:"sha256"`
	Token    string `json:"file_token"`
}

type catalogFile struct {
	Revision int      `json:"revision"`
	Themes   []*Theme `json:"themes"`
}

type store struct {
	mu        sync.Mutex
	lastCheck time.Time
	catalog   catalogFile
}

var cached store

// Warm 启动时拉一次。授权暂时不可用时不阻断官网启动。
func Warm(ctx context.Context) {
	if err := Ensure(ctx, true); err != nil {
		g.Log().Warningf(ctx, "底栏图标预热失败: %v", err)
	}
}

// Ensure 目录过期时向授权核对。force 忽略 10 分钟间隔。
func Ensure(ctx context.Context, force bool) error {
	cached.mu.Lock()
	defer cached.mu.Unlock()
	if err := loadLocked(ctx); err != nil {
		return err
	}
	if !force && !cached.lastCheck.IsZero() && time.Since(cached.lastCheck) < refreshInterval {
		return nil
	}
	remote, err := fetchCatalog(ctx)
	if err != nil {
		if cached.catalog.Revision > 0 || len(cached.catalog.Themes) > 0 {
			cached.lastCheck = time.Now()
			g.Log().Warningf(ctx, "底栏图标目录刷新失败，继续用本地缓存: %v", err)
			return nil
		}
		return err
	}
	if remote.Revision == cached.catalog.Revision && sameThemes(cached.catalog.Themes, remote.Themes) {
		cached.lastCheck = time.Now()
		return nil
	}
	if err = syncThemes(ctx, remote); err != nil {
		if cached.catalog.Revision > 0 || len(cached.catalog.Themes) > 0 {
			cached.lastCheck = time.Now()
			g.Log().Warningf(ctx, "底栏图标同步失败，继续用本地缓存: %v", err)
			return nil
		}
		return err
	}
	cached.catalog = *remote
	cached.lastCheck = time.Now()
	return saveLocked(ctx)
}

const appPageSize = 5

// appPageWindow 每页固定 5 套。内置「经典」不占目录页。
func appPageWindow(page int) (start, limit int) {
	if page < 1 {
		page = 1
	}
	return (page - 1) * appPageSize, appPageSize
}

func catalogIndex(themes []*Theme) []map[string]any {
	index := make([]map[string]any, 0, len(themes))
	for _, theme := range themes {
		index = append(index, map[string]any{
			"id":       theme.ThemeKey,
			"name":     theme.Name,
			"summary":  theme.Summary,
			"sort":     theme.Sort,
			"revision": theme.Revision,
		})
	}
	return index
}

// Payload 给 App 的一页目录。只带这一页的预览图。
func Payload(ctx context.Context, clientRevision, page, pageSize int) (map[string]any, error) {
	if page < 1 {
		page = 1
	}
	_ = pageSize
	start, limit := appPageWindow(page)
	if err := Ensure(ctx, false); err != nil {
		return nil, err
	}
	cached.mu.Lock()
	defer cached.mu.Unlock()
	total := len(cached.catalog.Themes)
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	hasMore := end < total
	index := catalogIndex(cached.catalog.Themes)
	if clientRevision > 0 && clientRevision == cached.catalog.Revision && page == 1 {
		return map[string]any{
			"code":      0,
			"revision":  cached.catalog.Revision,
			"unchanged": true,
			"page":      page,
			"page_size": limit,
			"total":     total,
			"has_more":  hasMore,
			"index":     index,
		}, nil
	}
	themes := make([]map[string]any, 0, end-start)
	for _, theme := range cached.catalog.Themes[start:end] {
		icons, err := readIcons(ctx, theme)
		if err != nil {
			g.Log().Warningf(ctx, "读取底栏图标 %s 失败: %v", theme.ThemeKey, err)
			continue
		}
		themes = append(themes, map[string]any{
			"id":       theme.ThemeKey,
			"name":     theme.Name,
			"summary":  theme.Summary,
			"sort":     theme.Sort,
			"revision": theme.Revision,
			"sha256":   theme.SHA256,
			"icons":    icons,
		})
	}
	return map[string]any{
		"code":      0,
		"revision":  cached.catalog.Revision,
		"page":      page,
		"page_size": limit,
		"total":     total,
		"has_more":  end < total,
		"themes":    themes,
		"index":     index,
	}, nil
}

// ZipPayload 单套原图 zip。列表预览用压缩图，App 点下载时才来取这一份。
func ZipPayload(ctx context.Context, themeKey string, revision int) (map[string]any, error) {
	if err := Ensure(ctx, false); err != nil {
		return nil, err
	}
	cached.mu.Lock()
	defer cached.mu.Unlock()
	var theme *Theme
	for _, item := range cached.catalog.Themes {
		if item.ThemeKey == themeKey && (revision == 0 || item.Revision == revision) {
			theme = item
			break
		}
	}
	if theme == nil {
		return nil, fmt.Errorf("主题不存在")
	}
	zipPath := zipPath(ctx, theme)
	raw, err := os.ReadFile(zipPath)
	if err != nil {
		return nil, fmt.Errorf("图标包不存在")
	}
	return map[string]any{
		"code":     0,
		"id":       theme.ThemeKey,
		"revision": theme.Revision,
		"sha256":   theme.SHA256,
		"zip":      base64.StdEncoding.EncodeToString(raw),
	}, nil
}

// FilePack 原图包不进授权。授权只签一把解开它的钥匙，避免原图把签名请求撑过网关上限。
type FilePack struct {
	Meta []byte
	File []byte
}

func PackFile(ctx context.Context, themeKey string, revision int) (FilePack, error) {
	payload, err := ZipPayload(ctx, themeKey, revision)
	if err != nil {
		return FilePack{}, err
	}
	encoded, _ := payload["zip"].(string)
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) == 0 {
		return FilePack{}, fmt.Errorf("图标包不存在")
	}
	delete(payload, "zip")
	key := make([]byte, 32)
	nonce := make([]byte, 12)
	if _, err = rand.Read(key); err != nil {
		return FilePack{}, err
	}
	if _, err = rand.Read(nonce); err != nil {
		return FilePack{}, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return FilePack{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return FilePack{}, err
	}
	payload["file_key"] = base64.StdEncoding.EncodeToString(key)
	payload["file_nonce"] = base64.StdEncoding.EncodeToString(nonce)
	meta, err := json.Marshal(payload)
	if err != nil {
		return FilePack{}, err
	}
	return FilePack{Meta: meta, File: gcm.Seal(nil, nonce, raw, nil)}, nil
}

// PackPreview 把带预览图的目录正文留在官网。授权只签钥匙，重的一页才不会被网关拒绝。
func PackPreview(payload map[string]any) (FilePack, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return FilePack{}, err
	}
	key := make([]byte, 32)
	nonce := make([]byte, 12)
	if _, err = rand.Read(key); err != nil {
		return FilePack{}, err
	}
	if _, err = rand.Read(nonce); err != nil {
		return FilePack{}, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return FilePack{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return FilePack{}, err
	}
	meta, err := json.Marshal(map[string]any{
		"code":          payload["code"],
		"preview_key":   base64.StdEncoding.EncodeToString(key),
		"preview_nonce": base64.StdEncoding.EncodeToString(nonce),
	})
	if err != nil {
		return FilePack{}, err
	}
	return FilePack{Meta: meta, File: gcm.Seal(nil, nonce, raw, nil)}, nil
}

func readIcons(ctx context.Context, theme *Theme) (map[string]map[string]string, error) {
	dir := themeDir(ctx, theme)
	out := map[string]map[string]string{}
	for _, slot := range slots {
		normal, err := os.ReadFile(filepath.Join(dir, slot+".png"))
		if err != nil {
			return nil, err
		}
		selected, err := os.ReadFile(filepath.Join(dir, slot+"_selected.png"))
		if err != nil {
			return nil, err
		}
		normal, err = previewPNG(normal)
		if err != nil {
			return nil, err
		}
		selected, err = previewPNG(selected)
		if err != nil {
			return nil, err
		}
		out[slot] = map[string]string{
			"normal":   base64.StdEncoding.EncodeToString(normal),
			"selected": base64.StdEncoding.EncodeToString(selected),
		}
	}
	return out, nil
}

// previewEdge 只限制发给 App 的预览图，磁盘上的原图不动。
// 一页 5 套如果按 256 原图签回包，会超过授权网关的请求体上限。
const previewEdge = 80

func previewPNG(data []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	edge := w
	if h > edge {
		edge = h
	}
	if edge <= previewEdge {
		return data, nil
	}
	nw := w * previewEdge / edge
	nh := h * previewEdge / edge
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		y0 := bounds.Min.Y + y*h/nh
		y1 := bounds.Min.Y + (y+1)*h/nh
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < nw; x++ {
			x0 := bounds.Min.X + x*w/nw
			x1 := bounds.Min.X + (x+1)*w/nw
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var r, g, b, a, n uint32
			for yy := y0; yy < y1; yy++ {
				for xx := x0; xx < x1; xx++ {
					cr, cg, cb, ca := img.At(xx, yy).RGBA()
					r += cr
					g += cg
					b += cb
					a += ca
					n++
				}
			}
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8((r / n) >> 8),
				G: uint8((g / n) >> 8),
				B: uint8((b / n) >> 8),
				A: uint8((a / n) >> 8),
			})
		}
	}
	var buf bytes.Buffer
	if err = png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func syncThemes(ctx context.Context, remote *catalogFile) error {
	keep := map[string]struct{}{}
	for _, theme := range remote.Themes {
		keep[theme.ThemeKey] = struct{}{}
		if localSHA(ctx, theme) == strings.ToLower(theme.SHA256) && iconsReady(ctx, theme) {
			continue
		}
		raw, err := downloadZip(ctx, theme)
		if err != nil {
			return err
		}
		got := sha256.Sum256(raw)
		if !strings.EqualFold(hex.EncodeToString(got[:]), theme.SHA256) {
			return fmt.Errorf("主题 %s SHA256 不匹配", theme.ThemeKey)
		}
		dir := themeDir(ctx, theme)
		if err = os.RemoveAll(dir); err != nil {
			return err
		}
		if err = gfile.Mkdir(dir); err != nil {
			return err
		}
		if err = extractIcons(raw, dir); err != nil {
			return err
		}
		if err = os.WriteFile(zipPath(ctx, theme), raw, 0o644); err != nil {
			return err
		}
	}
	root := cacheDir(ctx)
	entries, _ := os.ReadDir(root)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, ok := keep[entry.Name()]; ok {
			continue
		}
		_ = os.RemoveAll(filepath.Join(root, entry.Name()))
	}
	return nil
}

func extractIcons(raw []byte, dir string) error {
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return err
	}
	want := map[string]struct{}{}
	for _, slot := range slots {
		want[slot+".png"] = struct{}{}
		want[slot+"_selected.png"] = struct{}{}
	}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		base := path.Base(strings.ReplaceAll(file.Name, "\\", "/"))
		if _, ok := want[base]; !ok {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(io.LimitReader(rc, 512*1024))
		rc.Close()
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(dir, base), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func iconsReady(ctx context.Context, theme *Theme) bool {
	dir := themeDir(ctx, theme)
	for _, slot := range slots {
		if !gfile.Exists(filepath.Join(dir, slot+".png")) || !gfile.Exists(filepath.Join(dir, slot+"_selected.png")) {
			return false
		}
	}
	return true
}

func localSHA(ctx context.Context, theme *Theme) string {
	raw, err := os.ReadFile(zipPath(ctx, theme))
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func fetchCatalog(ctx context.Context) (*catalogFile, error) {
	const pageSize = 50
	var all []*Theme
	revision := 0
	for page := 1; page < 500; page++ {
		var envelope struct {
			Code int `json:"code"`
			Data struct {
				Revision int      `json:"revision"`
				Total    int      `json:"total"`
				Themes   []*Theme `json:"themes"`
			} `json:"data"`
			Message string `json:"message"`
		}
		body := map[string]any{"page": page, "page_size": pageSize}
		if err := postAuth(ctx, "/api/v1/sign/app/tab-icon/catalog", body, &envelope); err != nil {
			return nil, err
		}
		if envelope.Code != 0 {
			if envelope.Message == "" {
				envelope.Message = "授权目录失败"
			}
			return nil, fmt.Errorf("%s", envelope.Message)
		}
		if page == 1 {
			revision = envelope.Data.Revision
		}
		if envelope.Data.Themes == nil {
			break
		}
		all = append(all, envelope.Data.Themes...)
		if len(envelope.Data.Themes) < pageSize || (envelope.Data.Total > 0 && len(all) >= envelope.Data.Total) {
			break
		}
	}
	if all == nil {
		all = []*Theme{}
	}
	return &catalogFile{Revision: revision, Themes: all}, nil
}

func downloadZip(ctx context.Context, theme *Theme) ([]byte, error) {
	base, key, err := authConfig(ctx)
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(map[string]any{"id": theme.ID, "token": theme.Token})
	cipher, err := siteaes.Encrypt(key, string(body))
	if err != nil {
		return nil, err
	}
	wrap, _ := json.Marshal(map[string]string{"cipher": cipher})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v1/sign/app/tab-icon/file", bytes.NewReader(wrap))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 2 * time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(resp.Header.Get("Content-Type"), "json") || resp.StatusCode >= 400 {
		plain, decErr := decodeAuthBody(key, raw)
		if decErr == nil {
			var envelope struct {
				Message string `json:"message"`
			}
			_ = json.Unmarshal(plain, &envelope)
			if envelope.Message != "" {
				return nil, fmt.Errorf("%s", envelope.Message)
			}
		}
		return nil, fmt.Errorf("下载图标包失败: HTTP %d", resp.StatusCode)
	}
	return raw, nil
}

// Opened 是授权拆开 App 信封后，官网组包要用的字段。
type Opened struct {
	Revision int
	ThemeKey string
	Page     int
	PageSize int
}

// OpenEnvelope 把 App 信封交给授权拆开，只取目录修订号、主题键和页码。回包私钥留在授权。
func OpenEnvelope(ctx context.Context, envelope json.RawMessage) (Opened, error) {
	var data struct {
		Code int `json:"code"`
		Data struct {
			Revision int    `json:"revision"`
			ThemeKey string `json:"theme_key"`
			Page     int    `json:"page"`
			PageSize int    `json:"page_size"`
		} `json:"data"`
		Message string `json:"message"`
	}
	body := map[string]json.RawMessage{"envelope": envelope}
	if err := postAuth(ctx, "/api/v1/sign/app/tab-icon/open", body, &data); err != nil {
		return Opened{}, err
	}
	if data.Code != 0 {
		if data.Message == "" {
			data.Message = "拆信封失败"
		}
		return Opened{}, fmt.Errorf("%s", data.Message)
	}
	page := data.Data.Page
	if page < 1 {
		page = 1
	}
	pageSize := data.Data.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	return Opened{
		Revision: data.Data.Revision,
		ThemeKey: data.Data.ThemeKey,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// SealReply 把组好的正文交回授权，用授权已有的私钥签成 App 能解开的回包。
func SealReply(ctx context.Context, envelope json.RawMessage, payload []byte) (map[string]any, error) {
	var data struct {
		Code int `json:"code"`
		Data struct {
			V          int    `json:"v"`
			IV         string `json:"iv"`
			Ciphertext string `json:"ciphertext"`
			Tag        string `json:"tag"`
			Sig        string `json:"sig"`
		} `json:"data"`
		Message string `json:"message"`
	}
	body := map[string]json.RawMessage{
		"envelope": envelope,
		"payload":  json.RawMessage(payload),
	}
	if err := postAuth(ctx, "/api/v1/sign/app/tab-icon/seal", body, &data); err != nil {
		return nil, err
	}
	if data.Code != 0 {
		if data.Message == "" {
			data.Message = "密封失败"
		}
		return nil, fmt.Errorf("%s", data.Message)
	}
	if data.Data.Ciphertext == "" || data.Data.Sig == "" {
		return nil, fmt.Errorf("密封失败")
	}
	return map[string]any{
		"v":          data.Data.V,
		"iv":         data.Data.IV,
		"ciphertext": data.Data.Ciphertext,
		"tag":        data.Data.Tag,
		"sig":        data.Data.Sig,
	}, nil
}

func postAuth(ctx context.Context, apiPath string, payload any, out any) error {
	base, key, err := authConfig(ctx)
	if err != nil {
		return err
	}
	plain, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	status, body, err := postEncrypted(ctx, base+apiPath, key, plain)
	if err != nil {
		return err
	}
	if status >= 400 && len(body) == 0 {
		return fmt.Errorf("授权接口 HTTP %d", status)
	}
	if err = json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("解析授权响应失败: %w", err)
	}
	return nil
}

func authConfig(ctx context.Context) (base, key string, err error) {
	base = strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authBaseUrl").String()), "/")
	key = strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.authAesKey").String())
	if base == "" || key == "" {
		return "", "", fmt.Errorf("未配置授权地址")
	}
	return base, key, nil
}

func postEncrypted(ctx context.Context, api, key string, plain []byte) (int, []byte, error) {
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
	resp, err := (&http.Client{Timeout: 90 * time.Second}).Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	decoded, err := decodeAuthBody(key, body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, decoded, nil
}

func decodeAuthBody(key string, body []byte) ([]byte, error) {
	var wrap struct {
		Cipher string `json:"cipher"`
	}
	if json.Unmarshal(bytes.TrimSpace(body), &wrap) == nil && strings.TrimSpace(wrap.Cipher) != "" {
		plain, err := siteaes.Decrypt(key, wrap.Cipher)
		if err != nil {
			return nil, fmt.Errorf("授权响应解密失败: %w", err)
		}
		return []byte(plain), nil
	}
	return nil, fmt.Errorf("授权响应未加密")
}

func sameThemes(a, b []*Theme) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ThemeKey != b[i].ThemeKey || a[i].Revision != b[i].Revision || !strings.EqualFold(a[i].SHA256, b[i].SHA256) {
			return false
		}
	}
	return true
}

func cacheDir(ctx context.Context) string {
	dir := strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.tabIconCacheDir").String())
	if dir == "" {
		dir = "resource/base/tab-icons"
	}
	return dir
}

func statePath(ctx context.Context) string {
	return filepath.Join(cacheDir(ctx), "state.json")
}

func themeDir(ctx context.Context, theme *Theme) string {
	return filepath.Join(cacheDir(ctx), theme.ThemeKey, fmt.Sprintf("%d", theme.Revision))
}

func zipPath(ctx context.Context, theme *Theme) string {
	return filepath.Join(cacheDir(ctx), theme.ThemeKey, fmt.Sprintf("%d.zip", theme.Revision))
}

func loadLocked(ctx context.Context) error {
	if cached.catalog.Revision > 0 || cached.catalog.Themes != nil {
		return nil
	}
	raw, err := os.ReadFile(statePath(ctx))
	if err != nil {
		if os.IsNotExist(err) {
			cached.catalog.Themes = []*Theme{}
			return nil
		}
		return err
	}
	return json.Unmarshal(raw, &cached.catalog)
}

func saveLocked(ctx context.Context) error {
	if err := gfile.Mkdir(cacheDir(ctx)); err != nil {
		return err
	}
	raw, err := json.Marshal(cached.catalog)
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(ctx), raw, 0o644)
}
