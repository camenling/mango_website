package apptheme

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/util/guid"
)

const (
	MinID = 1
	MaxID = 5

	// MaxPackageBytes 外部上传 .hmtheme 的大小上限，与 frontend/src/config/hmTheme.js 的 HMTHEME_MAX_BYTES 一致
	MaxPackageBytes = 30 << 20
)

// ErrBadPackage 上传的主题包不是合法 .hmtheme
var ErrBadPackage = gerror.New("bad theme package")

// AccentHexByTheme 与 frontend/src/config/themes.js 五套强调色一致
var AccentHexByTheme = map[int]string{
	1: "#0a84ff", // 沧澜
	2: "#d97a55", // 蜜晖
	3: "#30b0c7", // 青岚
	4: "#ff375f", // 绛霞
	5: "#c9922e", // 流金
}

var nameByTheme = map[int]string{
	1: "沧澜",
	2: "蜜晖",
	3: "青岚",
	4: "绛霞",
	5: "流金",
}

// Config 签名时内置的主题：风格、强调色、玻璃渐变开关
type Config struct {
	Theme  int
	Accent string
	Glass  bool
}

// ApplyResult 预设主题应用结果（供签名日志）
type ApplyResult struct {
	Theme  int
	Accent string
	Glass  bool
}

var accentPattern = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

// NormalizeAccent 规范强调色为 #RRGGBB；空或不合法时用风格默认色
func NormalizeAccent(theme int, accent string) string {
	accent = strings.TrimSpace(accent)
	if !accentPattern.MatchString(accent) {
		return strings.ToUpper(AccentHex(theme))
	}
	return "#" + strings.ToUpper(strings.TrimPrefix(accent, "#"))
}

// NormalizeID 规范主题编号（1–5，对应 HillMountPlatform HMAppIconOption）
func NormalizeID(theme int) int {
	if theme < MinID || theme > MaxID {
		return MinID
	}
	return theme
}

// AccentHex 返回风格默认强调色
func AccentHex(theme int) string {
	theme = NormalizeID(theme)
	if hex, ok := AccentHexByTheme[theme]; ok {
		return hex
	}
	return AccentHexByTheme[MinID]
}

// PublicIconPath 站点静态图标（OTA plist / 网页展示）
func PublicIconPath(theme int) string {
	return fmt.Sprintf("/themes/%d.png", NormalizeID(theme))
}

// builtinThemeRel HillMount 首次启动读取的内置主题方案（与 App 内导出的 .hmtheme 同一格式）
const builtinThemeRel = "website/hm-builtin-theme.hmtheme"

// manifest 字段与 HillMount HMThemeProfileManifest 一致；只带风格和字体颜色两个模块，其余模块保持 App 默认。
type manifest struct {
	SchemaVersion int                    `json:"schemaVersion"`
	AppVersion    string                 `json:"appVersion"`
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	CreatedAt     string                 `json:"createdAt"`
	UpdatedAt     string                 `json:"updatedAt"`
	Source        map[string]string      `json:"source"`
	Modules       map[string]interface{} `json:"modules"`
}

// BuildThemePackage 生成 .hmtheme（zip，根目录 manifest.json）
func BuildThemePackage(cfg Config) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	m := manifest{
		SchemaVersion: 1,
		AppVersion:    "",
		ID:            guid.S(),
		Name:          "官网预设 · " + nameByTheme[cfg.Theme],
		CreatedAt:     now,
		UpdatedAt:     now,
		Source:        map[string]string{"kind": "local"},
		Modules: map[string]interface{}{
			"style": map[string]interface{}{
				"v":     1,
				"style": cfg.Theme,
				"glass": cfg.Glass,
			},
			"textColor": map[string]interface{}{
				"v":      1,
				"colors": map[string]string{"accentLight": cfg.Accent},
			},
		},
	}
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("manifest.json")
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(body); err != nil {
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EmbedTheme 把主题方案写进 .app/website/。
// 不修改 .app 内主图标 PNG / Info.plist，保证 App 内 1–5 切换时主题 1 仍是原版沧澜（AppIcon 主集）。
func EmbedTheme(appPath string, cfg Config) error {
	data, err := BuildThemePackage(cfg)
	if err != nil {
		return gerror.Wrap(err, "生成内置主题失败")
	}
	if err := gfile.Mkdir(filepath.Join(appPath, "website")); err != nil {
		return gerror.Wrap(err, "创建 website 目录失败")
	}
	if err := gfile.PutBytes(filepath.Join(appPath, builtinThemeRel), data); err != nil {
		return gerror.Wrap(err, "写入 hm-builtin-theme.hmtheme 失败")
	}
	return nil
}

// ValidatePackage 校验外部 .hmtheme：zip 可读、大小在上限内、根目录有 manifest.json 且 schemaVersion 为 1。
func ValidatePackage(data []byte) error {
	if len(data) == 0 || len(data) > MaxPackageBytes {
		return ErrBadPackage
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ErrBadPackage
	}
	for _, f := range zr.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return ErrBadPackage
		}
		body, err := io.ReadAll(io.LimitReader(rc, 4<<20))
		rc.Close()
		if err != nil {
			return ErrBadPackage
		}
		var m struct {
			SchemaVersion int                    `json:"schemaVersion"`
			Modules       map[string]interface{} `json:"modules"`
		}
		if err = json.Unmarshal(body, &m); err != nil || m.SchemaVersion != 1 || len(m.Modules) == 0 {
			return ErrBadPackage
		}
		return nil
	}
	return ErrBadPackage
}

// EmbedPackage 把外部生成的 .hmtheme（主题配置页导出）原样写进 .app/website/。
func EmbedPackage(ctx context.Context, appPath string, data []byte) error {
	if err := ValidatePackage(data); err != nil {
		return err
	}
	if err := gfile.Mkdir(filepath.Join(appPath, "website")); err != nil {
		return gerror.Wrap(err, "创建 website 目录失败")
	}
	if err := gfile.PutBytes(filepath.Join(appPath, builtinThemeRel), data); err != nil {
		return gerror.Wrap(err, "写入 hm-builtin-theme.hmtheme 失败")
	}
	g.Log().Infof(ctx, "内置主题包 %d 字节已写入 %s", len(data), builtinThemeRel)
	return nil
}

// ApplyTheme 签名前内置主题；图标由 HillMount 内 setAlternateIconName 切换，主图标 AppIcon 保持不动。
func ApplyTheme(ctx context.Context, appPath string, cfg Config) (*ApplyResult, error) {
	cfg.Theme = NormalizeID(cfg.Theme)
	cfg.Accent = NormalizeAccent(cfg.Theme, cfg.Accent)
	if err := EmbedTheme(appPath, cfg); err != nil {
		return nil, err
	}
	g.Log().Infof(ctx,
		"内置主题 theme=%d accent=%s glass=%t 已写入 %s（主图标 AppIcon 未改动，App 内仍保留 1–5 完整切换）",
		cfg.Theme, cfg.Accent, cfg.Glass, builtinThemeRel,
	)
	return &ApplyResult{Theme: cfg.Theme, Accent: cfg.Accent, Glass: cfg.Glass}, nil
}
