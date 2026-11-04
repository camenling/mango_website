package install

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
)

const (
	shortPlistPrefix     = "preinstall_install_short:"
	shortPlistHashPrefix = "preinstall_install_hash:"
	shortKeyLen          = 6
	shortKeyChars        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var shortCache = gcache.New()

func shortTTL(ctx context.Context) time.Duration {
	hours := g.Cfg().MustGet(ctx, "sign.shortLinkExpireHours").Int()
	if hours <= 0 {
		hours = 24
	}
	return time.Duration(hours) * time.Hour
}

func cacheSet(ctx context.Context, key, val string, ttl time.Duration) error {
	if useRedis(ctx) {
		_, err := g.Redis().Do(ctx, "SET", key, val, "EX", int(ttl.Seconds()))
		return err
	}
	return shortCache.Set(ctx, key, val, ttl)
}

func cacheGet(ctx context.Context, key string) (string, error) {
	if useRedis(ctx) {
		v, err := g.Redis().Do(ctx, "GET", key)
		if err != nil {
			return "", err
		}
		if v == nil || v.IsNil() {
			return "", nil
		}
		return v.String(), nil
	}
	v, err := shortCache.Get(ctx, key)
	if err != nil || v.IsNil() {
		return "", err
	}
	return v.String(), nil
}

func useRedis(ctx context.Context) bool {
	cfg, err := g.Cfg().Get(ctx, "redis.default.address")
	return err == nil && cfg != nil && strings.TrimSpace(cfg.String()) != ""
}

// FindOrCreateShortInstallPlistKey 为 plist URL 创建 /i/ 短链 key（对齐 sign/backend）
func FindOrCreateShortInstallPlistKey(ctx context.Context, plistURL string) (string, error) {
	plistURL = strings.TrimSpace(plistURL)
	if plistURL == "" {
		return "", errors.New("plist 地址不能为空")
	}
	hashKey := shortPlistHashPrefix + fmt.Sprintf("%x", md5.Sum([]byte(plistURL)))
	if existing, _ := cacheGet(ctx, hashKey); existing != "" {
		return existing, nil
	}
	key := genRandomKey(shortKeyLen)
	if key == "" {
		return "", errors.New("生成短 key 失败")
	}
	ttl := shortTTL(ctx)
	if err := cacheSet(ctx, shortPlistPrefix+key, plistURL, ttl); err != nil {
		return "", err
	}
	_ = cacheSet(ctx, hashKey, key, ttl)
	return key, nil
}

// GetShortInstallPlist 解析 /i/:key 对应的 plist URL
func GetShortInstallPlist(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", errors.New("key 不能为空")
	}
	val, err := cacheGet(ctx, shortPlistPrefix+key)
	if err != nil {
		return "", err
	}
	if val == "" {
		return "", errors.New("链接已失效或不存在")
	}
	return val, nil
}

func genRandomKey(n int) string {
	b := make([]byte, n)
	charsLen := big.NewInt(int64(len(shortKeyChars)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, charsLen)
		if err != nil {
			return ""
		}
		b[i] = shortKeyChars[idx.Int64()]
	}
	return string(b)
}
