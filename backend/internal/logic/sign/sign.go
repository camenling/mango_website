package sign

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
	"howett.net/plist"

	"mangosign/internal/logic/zsign"
	"mangosign/internal/model"
	"mangosign/internal/pkg/appsource"
	"mangosign/internal/pkg/apptheme"
	"mangosign/internal/pkg/baseipa"
	"mangosign/internal/pkg/certmango"
	installpkg "mangosign/internal/pkg/install"
	"mangosign/internal/service"
)

type sSign struct{}

func init() {
	service.RegisterSign(New())
}

func New() *sSign {
	return &sSign{}
}

func (s *sSign) SignAndInstall(ctx context.Context, in *model.SignInput) (*model.SignResult, error) {
	cfg := g.Cfg()
	baseIpa, rel, err := baseipa.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	if !gfile.Exists(baseIpa) {
		return nil, gerror.New("签名底包不存在")
	}
	if rel != nil {
		g.Log().Infof(ctx, "签名底包使用已发布版本 %s / %s -> %s", rel.HmBuildDate, rel.Version, baseIpa)
	} else {
		g.Log().Infof(ctx, "签名底包回退本地文件 %s", baseIpa)
	}

	taskId := gfile.Name(gfile.Dir(in.P12Path))
	tempDir := gfile.Dir(in.P12Path)
	outputDir := cfg.MustGet(ctx, "sign.outputDir").String()
	signedDir := gfile.Join(outputDir, taskId)
	if err := gfile.Mkdir(signedDir); err != nil {
		return nil, gerror.Wrap(err, "创建输出目录失败")
	}

	outIpa := gfile.Join(signedDir, "app.ipa")
	appName := cfg.MustGet(ctx, "sign.appName").String()
	appVersion := "1.0.0"
	if rel != nil && strings.TrimSpace(rel.Version) != "" {
		appVersion = strings.TrimSpace(rel.Version)
	}

	themeID := apptheme.NormalizeID(in.Theme)
	appSource := strings.TrimSpace(in.AppSource)
	if _, err := appsource.NormalizeURLs(appSource); err != nil {
		return nil, err
	}

	bundleId, err := s.signIpaWithParam(ctx, signIpaParam{
		TaskID:              taskId,
		IpaPath:             baseIpa,
		P12Path:             in.P12Path,
		MobileprovisionPath: in.MobileprovisionPath,
		Password:            in.Password,
		OutputPath:          outIpa,
		AppName:             appName,
		AppBundleID:         "", // 保留底包 CFBundleIdentifier，不按描述文件改 Bundle
		AppVersion:          appVersion,
		Theme:               themeID,
		Accent:              in.Accent,
		Glass:               in.Glass,
		AppSource:           appSource,
	})
	if err != nil {
		return nil, err
	}

	siteUrl := strings.TrimRight(cfg.MustGet(ctx, "sign.siteUrl").String(), "/")
	ipaUrl := fmt.Sprintf("%s/signed/%s/app.ipa", siteUrl, taskId)
	iconUrl := siteUrl + apptheme.PublicIconPath(themeID)

	plUUID, err := installpkg.SavePlistInfo(ctx, model.InstallPlist{
		Name:       appName,
		BundleID:   bundleId,
		Version:    appVersion,
		IconUrl:    iconUrl,
		AppFileUrl: ipaUrl,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "保存安装信息失败")
	}
	plistUrl := fmt.Sprintf("%s/getInstallPlist/%s", siteUrl, plUUID)
	itmsUrl := installpkg.ItmsInstallURL(plistUrl)

	installUrl := itmsUrl
	if cfg.MustGet(ctx, "sign.useShortInstall").Bool() {
		shortKey, sErr := installpkg.FindOrCreateShortInstallPlistKey(ctx, plistUrl)
		if sErr != nil {
			g.Log().Warningf(ctx, "短链生成失败，回退 itms 链接: %v", sErr)
		} else {
			installUrl = fmt.Sprintf("%s/i/%s", siteUrl, shortKey)
		}
	}

	// 兼容旧路径：静态 manifest（可选）
	if err := s.writeManifestPlist(signedDir, appName, bundleId, appVersion, ipaUrl); err != nil {
		return nil, gerror.Wrap(err, "生成安装描述文件失败")
	}

	if cfg.MustGet(ctx, "sign.cleanupCerts").Bool() {
		gfile.Remove(in.P12Path)
		gfile.Remove(in.MobileprovisionPath)
		gfile.Remove(tempDir)
	}

	return &model.SignResult{
		TaskId:     taskId,
		InstallUrl: installUrl,
		PlistUrl:   plistUrl,
		IpaUrl:     ipaUrl,
		ItmsUrl:    itmsUrl,
		Message:    "签名完成",
	}, nil
}

type signIpaParam struct {
	TaskID              string
	IpaPath             string
	P12Path             string
	MobileprovisionPath string
	Password            string
	OutputPath          string
	AppName             string
	AppBundleID         string
	AppVersion          string
	Theme               int
	Accent              string
	Glass               bool
	AppSource           string
}

// signIpaWithParam 解压 IPA → libzsign(CGO) 签名，流程对齐 sign/backend。
// 返回实际用于 OTA 的 Bundle ID（来自底包 Info.plist，除非 AppBundleID 非空）。
func (s *sSign) signIpaWithParam(ctx context.Context, param signIpaParam) (resolvedBundleID string, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		workDir := gfile.Dir(param.OutputPath)
		unzipDir := gfile.Join(workDir, "app")
		if err := gfile.Mkdir(unzipDir); err != nil {
			panic(gerror.Wrap(err, "创建解压目录失败"))
		}

		if err := unzipIPA(param.IpaPath, unzipDir); err != nil {
			panic(gerror.Wrap(err, "解压 IPA 失败"))
		}

		appUnzipPath, err := findAppInPayload(unzipDir)
		if err != nil {
			panic(gerror.Wrap(err, "查找 .app 失败"))
		}

		appName, appBundleID, appVersion := param.AppName, param.AppBundleID, param.AppVersion
		appInfoPlistPath := filepath.Join(appUnzipPath, "Info.plist")
		if gfile.Exists(appInfoPlistPath) {
			readName, readBundle, readVer, e := readInfoPlistFields(appInfoPlistPath)
			if e == nil {
				if appName == "" {
					appName = readName
				}
				if appBundleID == "" {
					appBundleID = readBundle
				}
				if appVersion == "" {
					appVersion = readVer
				}
			}
		}

		if _, err := apptheme.ApplyTheme(ctx, appUnzipPath, apptheme.Config{
			Theme:  param.Theme,
			Accent: param.Accent,
			Glass:  param.Glass,
		}); err != nil {
			panic(gerror.Wrap(err, "应用预设主题失败"))
		}

		if err := appsource.Embed(ctx, appUnzipPath, param.AppSource); err != nil {
			panic(gerror.Wrap(err, "内置软件源失败"))
		}

		if err := certmango.Embed(ctx, appUnzipPath, param.P12Path, param.MobileprovisionPath, param.Password); err != nil {
			panic(gerror.Wrap(err, "内置证书失败"))
		}
		g.Log().Infof(ctx, "内置证书已写入 %s/certificate/cert.mango", appUnzipPath)

		curDir, err := os.Getwd()
		if err != nil {
			panic(gerror.Wrap(err, "获取工作目录失败"))
		}

		zsignTempDir := taskZsignTempDir(ctx, param.TaskID)
		if err := gfile.Mkdir(zsignTempDir); err != nil {
			panic(gerror.Wrap(err, "创建 zsign 临时目录失败"))
		}
		defer gfile.Remove(zsignTempDir)

		signRet, err := zsign.Sign(ctx, zsign.SignParam{
			IpaPath:       filepath.Join(curDir, unzipDir),
			P12Path:       filepath.Join(curDir, param.P12Path),
			ProvisionPath: filepath.Join(curDir, param.MobileprovisionPath),
			OutputPath:    filepath.Join(curDir, param.OutputPath),
			CertPassword:  param.Password,
			BundleID:      appBundleID,
			BundleName:    appName,
			BundleVersion: appVersion,
			TempFolder:    filepath.Join(curDir, zsignTempDir),
		})
		if err != nil {
			panic(gerror.Wrap(err, "调用 zsign 失败"))
		}
		if gconv.Int(signRet["code"]) != 0 {
			msg := gconv.String(signRet["message"])
			panic(gerror.New("签名失败: " + msg))
		}

		outIpaAbs := filepath.Join(curDir, param.OutputPath)
		if !ipaContainsCertMango(outIpaAbs) {
			panic(gerror.New("签名完成但包内没有 certificate/cert.mango，内置证书未写入"))
		}

		resolvedBundleID = strings.TrimSpace(appBundleID)
		if outBundle, e := bundleIDFromIPA(filepath.Join(curDir, param.OutputPath)); e == nil && outBundle != "" {
			resolvedBundleID = outBundle
		}
		if resolvedBundleID == "" {
			panic(gerror.New("无法读取 IPA 的 Bundle ID，无法生成正确的 OTA 安装描述"))
		}
	})
	return resolvedBundleID, err
}

func findAppInPayload(unzipDir string) (string, error) {
	payloadDir := filepath.Join(unzipDir, "Payload")
	entries, err := os.ReadDir(payloadDir)
	if err != nil {
		return "", fmt.Errorf("读取 Payload 目录失败: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".app") {
			return filepath.Join(payloadDir, e.Name()), nil
		}
	}
	return "", gerror.New("未找到 .app 目录")
}

func ipaContainsCertMango(ipaPath string) bool {
	reader, err := zip.OpenReader(ipaPath)
	if err != nil {
		return false
	}
	defer reader.Close()
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := strings.ReplaceAll(f.Name, "\\", "/")
		if strings.HasSuffix(name, "/certificate/cert.mango") && f.UncompressedSize64 > 16 {
			return true
		}
	}
	return false
}

func taskZsignTempDir(ctx context.Context, taskID string) string {
	base := g.Cfg().MustGet(ctx, "sign.signTempDir").String()
	if base == "" {
		base = "resource/public/signTemp"
	}
	if taskID == "" {
		return base
	}
	return gfile.Join(base, taskID)
}

func readInfoPlistFields(plistPath string) (appName, bundleID, version string, err error) {
	return readInfoPlistBytes(gfile.GetBytes(plistPath))
}

func readInfoPlistBytes(data []byte) (appName, bundleID, version string, err error) {
	if len(data) == 0 {
		return "", "", "", gerror.New("Info.plist 为空")
	}
	var root map[string]interface{}
	if _, err = plist.Unmarshal(data, &root); err != nil {
		return "", "", "", err
	}
	appName = plistString(root, "CFBundleDisplayName")
	if appName == "" {
		appName = plistString(root, "CFBundleName")
	}
	bundleID = plistString(root, "CFBundleIdentifier")
	version = plistString(root, "CFBundleShortVersionString")
	if version == "" {
		version = plistString(root, "CFBundleVersion")
	}
	return appName, bundleID, version, nil
}

func plistString(root map[string]interface{}, key string) string {
	v, ok := root[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

// bundleIDFromIPA 从已签名 IPA 读取 CFBundleIdentifier，供 OTA manifest 与包内一致。
func bundleIDFromIPA(ipaPath string) (string, error) {
	reader, err := zip.OpenReader(ipaPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	for _, f := range reader.File {
		if f.FileInfo().IsDir() || !strings.HasPrefix(f.Name, "Payload/") {
			continue
		}
		if !strings.HasSuffix(f.Name, ".app/Info.plist") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", err
		}
		_, bundleID, _, err := readInfoPlistBytes(data)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(bundleID), nil
	}
	return "", gerror.New("IPA 内未找到 Info.plist")
}

func unzipIPA(ipaPath, extractPath string) error {
	reader, err := zip.OpenReader(ipaPath)
	if err != nil {
		return fmt.Errorf("打开 IPA 失败: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(extractPath, file.Name)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return fmt.Errorf("创建父目录失败: %w", err)
		}
		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开 zip 内文件失败: %w", err)
		}
		dst, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			src.Close()
			return fmt.Errorf("创建目标文件失败: %w", err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			src.Close()
			return fmt.Errorf("解压文件失败: %w", err)
		}
		dst.Close()
		src.Close()
	}
	return nil
}

func (s *sSign) writeManifestPlist(dir, appName, bundleId, appVersion, ipaUrl string) error {
	if strings.TrimSpace(bundleId) == "" {
		return gerror.New("Bundle ID 为空，无法生成 manifest.plist")
	}
	if appVersion == "" {
		appVersion = "1.0.0"
	}
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>items</key>
    <array>
        <dict>
            <key>assets</key>
            <array>
                <dict>
                    <key>kind</key>
                    <string>software-package</string>
                    <key>url</key>
                    <string>%s</string>
                </dict>
            </array>
            <key>metadata</key>
            <dict>
                <key>bundle-identifier</key>
                <string>%s</string>
                <key>bundle-version</key>
                <string>%s</string>
                <key>kind</key>
                <string>software</string>
                <key>title</key>
                <string>%s</string>
            </dict>
        </dict>
    </array>
</dict>
</plist>`, ipaUrl, bundleId, appVersion, appName)

	return gfile.PutContents(gfile.Join(dir, "manifest.plist"), content)
}

func (s *sSign) CheckCert(ctx context.Context, in *model.CertCheckInput) (*model.CertInfo, error) {
	info := &model.CertInfo{
		Status:  "unknown",
		Message: "",
	}

	cmd := exec.CommandContext(ctx, "openssl", "pkcs12", "-in", in.P12Path, "-passin", "pass:"+in.Password, "-nokeys", "-clcerts")
	out, err := cmd.CombinedOutput()
	if err != nil {
		info.Status = "invalid"
		info.Message = "证书密码错误或 P12 文件无效"
		return info, nil
	}

	outStr := string(out)
	if strings.Contains(outStr, "subject=") {
		info.Status = "valid"
	}
	if m := regexp.MustCompile(`O\s*=\s*([^,\n]+)`).FindStringSubmatch(outStr); len(m) > 1 {
		info.TeamName = strings.TrimSpace(m[1])
	}
	if m := regexp.MustCompile(`notAfter\s*=\s*([^\n]+)`).FindStringSubmatch(outStr); len(m) > 1 {
		info.ExpiryDate = strings.TrimSpace(m[1])
	}

	provisionData := gfile.GetContents(in.MobileprovisionPath)
	if provisionData != "" {
		if m := regexp.MustCompile(`<key>TeamName</key>\s*<string>([^<]+)</string>`).FindStringSubmatch(provisionData); len(m) > 1 {
			info.TeamName = strings.TrimSpace(m[1])
		}
		if m := regexp.MustCompile(`<key>TeamIdentifier</key>\s*<array>\s*<string>([^<]+)</string>`).FindStringSubmatch(provisionData); len(m) > 1 {
			info.TeamId = strings.TrimSpace(m[1])
		}
		if m := regexp.MustCompile(`<key>ExpirationDate</key>\s*<date>([^<]+)</date>`).FindStringSubmatch(provisionData); len(m) > 1 {
			info.ProvisionExpiry = strings.TrimSpace(m[1])
		}
		if m := regexp.MustCompile(`<key>Name</key>\s*<string>([^<]+)</string>`).FindStringSubmatch(provisionData); len(m) > 1 {
			info.AppName = strings.TrimSpace(m[1])
		}
		if m := regexp.MustCompile(`<key>application-identifier</key>\s*<string>([^<]+)</string>`).FindStringSubmatch(provisionData); len(m) > 1 {
			parts := strings.SplitN(strings.TrimSpace(m[1]), ".", 2)
			if len(parts) == 2 {
				info.BundleId = parts[1]
			}
		}
	}

	if info.Status == "valid" {
		info.Message = "证书解析成功"
	}
	return info, nil
}

func (s *sSign) StartCleanupWorker(ctx context.Context) {
	retention := g.Cfg().MustGet(ctx, "sign.tempRetentionMinutes").Int()
	if retention <= 0 {
		retention = 120
	}
	interval := g.Cfg().MustGet(ctx, "sign.cleanupIntervalMinutes").Int()
	if interval <= 0 {
		interval = 10
	}

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		defer ticker.Stop()
		s.cleanupOldFiles(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.cleanupOldFiles(ctx)
			}
		}
	}()
}

func (s *sSign) cleanupOldFiles(ctx context.Context) {
	cfg := g.Cfg()
	retention := cfg.MustGet(ctx, "sign.tempRetentionMinutes").Int()
	if retention <= 0 {
		retention = 120
	}
	cutoff := gtime.Now().Add(-time.Duration(retention) * time.Minute)

	for _, dir := range []string{
		cfg.MustGet(ctx, "sign.tempDir").String(),
		cfg.MustGet(ctx, "sign.outputDir").String(),
		cfg.MustGet(ctx, "sign.signTempDir").String(),
	} {
		s.removeExpiredTaskDirs(dir, cutoff.Time)
	}
}

// removeExpiredTaskDirs 删除根目录下「按 taskId 命名的子目录」中超过保留期的目录
func (s *sSign) removeExpiredTaskDirs(root string, cutoff time.Time) {
	if !gfile.Exists(root) {
		return
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := gfile.Join(root, e.Name())
		lastActive, err := taskDirLastActive(path)
		if err != nil {
			continue
		}
		if lastActive.Before(cutoff) {
			gfile.Remove(path)
		}
	}
}

// taskDirLastActive 判断目录是否过期：signed 以 app.ipa 为准；其它目录取直接子项最大 mtime（不递归进 Payload，避免 zip 内旧时间戳）
func taskDirLastActive(dir string) (time.Time, error) {
	if ipa := gfile.Join(dir, "app.ipa"); gfile.Exists(ipa) {
		if fi, err := gfile.Stat(ipa); err == nil {
			return fi.ModTime(), nil
		}
	}
	info, err := gfile.Stat(dir)
	if err != nil {
		return time.Time{}, err
	}
	latest := info.ModTime()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return latest, nil
	}
	for _, e := range entries {
		fi, err := e.Info()
		if err != nil {
			continue
		}
		if fi.ModTime().After(latest) {
			latest = fi.ModTime()
		}
	}
	return latest, nil
}

var _ = guid.S
