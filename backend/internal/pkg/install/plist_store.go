package install

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"mangosign/internal/model"
)

const plistInfoPrefix = "install_plist:"

func plistInfoTTL(ctx context.Context) time.Duration {
	hours := g.Cfg().MustGet(ctx, "sign.plistCacheExpireHours").Int()
	if hours <= 0 {
		hours = 2
	}
	return time.Duration(hours) * time.Hour
}

// SavePlistInfo 保存 OTA 安装信息，返回 getInstallPlist 的 uuid（对齐 sign/backend）
func SavePlistInfo(ctx context.Context, info model.InstallPlist) (uuid string, err error) {
	uuid = gconv.String(gtime.TimestampMilli())
	raw, err := json.Marshal(info)
	if err != nil {
		return "", fmt.Errorf("序列化 plist 信息失败: %w", err)
	}
	if err = cacheSet(ctx, plistInfoPrefix+uuid, string(raw), plistInfoTTL(ctx)); err != nil {
		return "", fmt.Errorf("缓存 plist 信息失败: %w", err)
	}
	return uuid, nil
}

// LoadPlistInfo 读取并生成 manifest 内容
func LoadPlistInfo(ctx context.Context, uuid string) (string, error) {
	if uuid == "" {
		return "", errors.New("uuid 不能为空")
	}
	raw, err := cacheGet(ctx, plistInfoPrefix+uuid)
	if err != nil {
		return "", err
	}
	if raw == "" {
		return "", errors.New("安装信息已过有效期")
	}
	var info model.InstallPlist
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return "", fmt.Errorf("解析 plist 信息失败: %w", err)
	}
	return MakeInstallPlist(info), nil
}
