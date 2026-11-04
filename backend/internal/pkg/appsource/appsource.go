package appsource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

const Rel = "website/hm-ota-appsource.json"

// Embed 把明文软件源入口写入 Payload/*.app/website/hm-ota-appsource.json。
// 不加密。raw 为空则跳过（保留底包里已有文件）。
func Embed(ctx context.Context, appPath, raw string) error {
	urls, err := NormalizeURLs(raw)
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		return nil
	}

	var payload []byte
	if len(urls) == 1 {
		payload, err = json.Marshal(map[string]string{"sourceURL": urls[0]})
	} else {
		payload, err = json.Marshal(map[string]any{"sources": urls})
	}
	if err != nil {
		return gerror.Wrap(err, "编码内置软件源失败")
	}

	websiteDir := filepath.Join(appPath, "website")
	if err := gfile.Mkdir(websiteDir); err != nil {
		return gerror.Wrap(err, "创建 website 目录失败")
	}
	dst := filepath.Join(appPath, Rel)
	if err := gfile.PutBytes(dst, payload); err != nil {
		return gerror.Wrap(err, "写入 hm-ota-appsource.json 失败")
	}
	g.Log().Infof(ctx, "内置软件源已写入 %s urls=%d", Rel, len(urls))
	return nil
}

func NormalizeURLs(raw string) ([]string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, nil
	}

	var ordered []string
	seen := map[string]struct{}{}
	add := func(item string) error {
		u, err := normalizeOne(item)
		if err != nil {
			return err
		}
		key := strings.ToLower(u)
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		ordered = append(ordered, u)
		return nil
	}

	if strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[") {
		var decoded any
		if err := json.Unmarshal([]byte(text), &decoded); err != nil {
			return nil, gerror.New("内置软件源 JSON 无效")
		}
		for _, item := range collectJSONURLs(decoded) {
			if err := add(item); err != nil {
				return nil, err
			}
		}
		if len(ordered) == 0 {
			return nil, gerror.New("内置软件源 JSON 里没有有效的 http/https 地址")
		}
		return ordered, nil
	}

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if err := add(line); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

func collectJSONURLs(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []any:
		var out []string
		for _, item := range t {
			out = append(out, collectJSONURLs(item)...)
		}
		return out
	case map[string]any:
		var out []string
		if sources, ok := t["sources"]; ok {
			out = append(out, collectJSONURLs(sources)...)
		}
		for _, key := range []string{"sourceURL", "sourceUrl"} {
			if s, ok := t[key].(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func normalizeOne(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, `"'`)
	if s == "" {
		return "", fmt.Errorf("软件源地址为空")
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", gerror.Newf("软件源地址不合法: %s", s)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", gerror.Newf("软件源仅支持 http/https: %s", s)
	}
	return s, nil
}
