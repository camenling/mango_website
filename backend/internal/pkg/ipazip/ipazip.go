package ipazip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Unzip 解压 IPA 到 dest，拒绝 zip slip。
func Unzip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("打开 IPA 失败: %w", err)
	}
	defer reader.Close()

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(destAbs, 0o755); err != nil {
		return err
	}

	for _, file := range reader.File {
		name := filepath.ToSlash(file.Name)
		if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "\\") {
			return gerror.New("zip 路径不合法")
		}
		target, err := safeJoin(destAbs, name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err = os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("创建目录失败: %w", err)
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("创建父目录失败: %w", err)
		}
		if err = extractFile(file, target); err != nil {
			return err
		}
	}
	return nil
}

// ZipDir 把 src 目录内容打成 IPA（保留 Payload 相对路径）。
func ZipDir(src, dest string) error {
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	w := zip.NewWriter(out)
	err = filepath.Walk(srcAbs, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, relErr := filepath.Rel(srcAbs, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		name := filepath.ToSlash(rel)
		if info.IsDir() {
			_, createErr := w.Create(name + "/")
			return createErr
		}
		hdr, hdrErr := zip.FileInfoHeader(info)
		if hdrErr != nil {
			return hdrErr
		}
		hdr.Name = name
		hdr.Method = zip.Deflate
		fw, createErr := w.CreateHeader(hdr)
		if createErr != nil {
			return createErr
		}
		f, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		_, copyErr := io.Copy(fw, f)
		f.Close()
		return copyErr
	})
	closeErr := w.Close()
	if err != nil {
		return err
	}
	return closeErr
}

// FindApp 返回 Payload 下第一个 .app。
func FindApp(unzipDir string) (string, error) {
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

func extractFile(file *zip.File, target string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开 zip 内文件失败: %w", err)
	}
	defer src.Close()
	dst, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("解压文件失败: %w", err)
	}
	return nil
}

func safeJoin(destAbs, name string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(name))
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return "", gerror.New("zip 路径不合法")
	}
	if filepath.IsAbs(cleaned) {
		return "", gerror.New("zip slip")
	}
	target := filepath.Join(destAbs, cleaned)
	rel, err := filepath.Rel(destAbs, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", gerror.New("zip slip")
	}
	return target, nil
}
