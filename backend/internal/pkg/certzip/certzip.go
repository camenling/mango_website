package certzip

import (
	"archive/zip"
	"io"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
)

const maxZipEntries = 200

// CertFiles 从压缩包或目录中识别出的证书文件。
type CertFiles struct {
	P12Path             string
	MobileprovisionPath string
	P12Name             string
	ProvisionName       string
}

// ExtractZip 解压 zip 到目标目录并识别证书文件。
func ExtractZip(zipPath, destDir string) (*CertFiles, error) {
	if err := gfile.Mkdir(destDir); err != nil {
		return nil, gerror.Wrap(err, "创建解压目录失败")
	}
	if err := unzip(zipPath, destDir); err != nil {
		return nil, err
	}
	return ScanDir(destDir)
}

// ScanDir 递归扫描目录，查找 p12/pfx 与 mobileprovision。
func ScanDir(dir string) (*CertFiles, error) {
	var p12Path, provisionPath string
	var p12Name, provisionName string

	entries, err := gfile.ScanDir(dir, "*", true)
	if err != nil {
		return nil, gerror.Wrap(err, "扫描目录失败")
	}

	for _, entry := range entries {
		if gfile.IsDir(entry) {
			continue
		}
		base := strings.ToLower(gfile.Basename(entry))
		switch {
		case strings.HasSuffix(base, ".p12") || strings.HasSuffix(base, ".pfx"):
			if p12Path == "" || strings.HasSuffix(base, ".p12") {
				p12Path = entry
				p12Name = gfile.Basename(entry)
			}
		case strings.HasSuffix(base, ".mobileprovision"):
			if provisionPath == "" {
				provisionPath = entry
				provisionName = gfile.Basename(entry)
			}
		}
	}

	if p12Path == "" {
		return nil, gerror.New("压缩包中未找到 .p12 或 .pfx 证书文件")
	}
	if provisionPath == "" {
		return nil, gerror.New("压缩包中未找到 .mobileprovision 描述文件")
	}

	return &CertFiles{
		P12Path:             p12Path,
		MobileprovisionPath: provisionPath,
		P12Name:             p12Name,
		ProvisionName:       provisionName,
	}, nil
}

func unzip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return gerror.Wrap(err, "无法打开 ZIP 文件")
	}
	defer reader.Close()

	if len(reader.File) > maxZipEntries {
		return gerror.New("压缩包内文件过多")
	}

	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return gerror.Wrap(err, "解析目标路径失败")
	}

	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		target, err := safeExtractPath(destAbs, f.Name)
		if err != nil {
			return err
		}
		if err = gfile.Mkdir(gfile.Dir(target)); err != nil {
			return gerror.Wrap(err, "创建解压子目录失败")
		}
		if err = extractFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func safeExtractPath(destDir, name string) (string, error) {
	clean := filepath.Clean(strings.ReplaceAll(name, "\\", "/"))
	if strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", gerror.New("压缩包包含非法路径")
	}
	target := filepath.Join(destDir, clean)
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", gerror.Wrap(err, "解析解压路径失败")
	}
	if !strings.HasPrefix(targetAbs, destDir+string(filepath.Separator)) && targetAbs != destDir {
		return "", gerror.New("压缩包包含非法路径")
	}
	return targetAbs, nil
}

func extractFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return gerror.Wrap(err, "读取压缩包条目失败")
	}
	defer rc.Close()

	out, err := gfile.Create(target)
	if err != nil {
		return gerror.Wrap(err, "写入解压文件失败")
	}
	defer out.Close()

	if _, err = io.Copy(out, rc); err != nil {
		return gerror.Wrap(err, "解压文件失败")
	}
	return nil
}
