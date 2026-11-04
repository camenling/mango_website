package certmango

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

var (
	reProfileUUID = regexp.MustCompile(`(?i)<key>UUID</key>\s*<string>([0-9a-f-]+)</string>`)
	reAllDevices  = regexp.MustCompile(`(?is)<key>ProvisionsAllDevices</key>\s*<true\s*/>`)
	reDevices     = regexp.MustCompile(`(?is)<key>ProvisionedDevices</key>\s*<array>(.*?)</array>`)
	rePlistString = regexp.MustCompile(`<string>([^<]+)</string>`)
)

// PlaceholderUDID 企业描述 / 多设备描述写入的占位，App 识别后只导入证书、不写入 UDID。
const PlaceholderUDID = "00000000-0000000000000000"

// Embed 把用户上传的 p12 + mobileprovision 加密成 Payload/*.app/certificate/cert.mango。
// 格式与 sign/backend、授权 unwrap 一致：AES-128-CBC，IV||密文，明文为
// UDID(48) + 密码(16) + 描述 UUID(48) + zip(p12+mp)。
func Embed(ctx context.Context, appUnzipPath, p12Path, provisionPath, p12Password string) error {
	if !gfile.Exists(p12Path) || !gfile.Exists(provisionPath) {
		return fmt.Errorf("p12 或 mobileprovision 不存在")
	}
	password := strings.TrimSpace(p12Password)
	if password == "" {
		return fmt.Errorf("证书密码不能为空")
	}
	if len(password) > 16 {
		return fmt.Errorf("证书密码超过 16 位，无法写入内置证书")
	}

	key, err := encryptKey(ctx)
	if err != nil {
		return err
	}

	mpContent := gfile.GetBytes(provisionPath)
	profileUUID, udid, err := parseProvision(mpContent)
	if err != nil {
		return err
	}

	zipData, err := packCertZip(udid, gfile.GetBytes(p12Path), mpContent)
	if err != nil {
		return fmt.Errorf("打包证书 zip 失败: %w", err)
	}

	blob, err := encrypt(zipData, udid, password, profileUUID, key)
	if err != nil {
		return err
	}

	certDir := filepath.Join(appUnzipPath, "certificate")
	if err := gfile.Mkdir(certDir); err != nil {
		return fmt.Errorf("创建证书目录失败: %w", err)
	}
	target := filepath.Join(certDir, "cert.mango")
	if err := gfile.PutBytes(target, blob); err != nil {
		return fmt.Errorf("写入 cert.mango 失败: %w", err)
	}
	return nil
}

func encryptKey(ctx context.Context) (string, error) {
	key := strings.TrimSpace(g.Cfg().MustGet(ctx, "sign.certMangoKey").String())
	if len(key) != 16 {
		return "", fmt.Errorf("sign.certMangoKey 必须是 16 字节，当前 %d", len(key))
	}
	return key, nil
}

func parseProvision(mp []byte) (profileUUID, udid string, err error) {
	src := provisionXML(mp)
	m := reProfileUUID.FindSubmatch(src)
	if len(m) < 2 {
		return "", "", fmt.Errorf("无法从描述文件提取 UUID")
	}
	profileUUID = strings.TrimSpace(string(m[1]))
	if profileUUID == "" {
		return "", "", fmt.Errorf("描述文件 UUID 为空")
	}

	if reAllDevices.Match(src) {
		return profileUUID, PlaceholderUDID, nil
	}

	var devices []string
	if block := reDevices.FindSubmatch(src); len(block) >= 2 {
		for _, sm := range rePlistString.FindAllSubmatch(block[1], -1) {
			if len(sm) < 2 {
				continue
			}
			id := strings.TrimSpace(string(sm[1]))
			if id != "" && !isPlaceholderUDID(id) {
				devices = append(devices, id)
			}
		}
	}
	if len(devices) == 1 {
		return profileUUID, devices[0], nil
	}
	return profileUUID, PlaceholderUDID, nil
}

func provisionXML(mp []byte) []byte {
	start := bytes.Index(mp, []byte("<?xml"))
	if start < 0 {
		start = bytes.Index(mp, []byte("<plist"))
	}
	if start < 0 {
		return mp
	}
	end := bytes.LastIndex(mp, []byte("</plist>"))
	if end < 0 || end < start {
		return mp[start:]
	}
	return mp[start : end+len("</plist>")]
}

func isPlaceholderUDID(udid string) bool {
	compact := strings.ReplaceAll(strings.TrimSpace(udid), "-", "")
	if compact == "" {
		return true
	}
	return strings.Trim(compact, "0") == ""
}

func packCertZip(udid string, p12, provision []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	if err := writeZipFile(zw, udid+".p12", p12); err != nil {
		zw.Close()
		return nil, err
	}
	if err := writeZipFile(zw, udid+".mobileprovision", provision); err != nil {
		zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func encrypt(zipData []byte, udid, password, profileUUID, key string) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("cert.mango 密钥必须是 16 字节")
	}
	plain := []byte(fmt.Sprintf("%-48s", udid) + fmt.Sprintf("%-16s", password) + fmt.Sprintf("%-48s", profileUUID))
	plain = append(plain, zipData...)
	plain = pkcs7Pad(plain, aes.BlockSize)

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("生成 IV 失败: %w", err)
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	ct := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ct, plain)
	return append(iv, ct...), nil
}

func decrypt(blob []byte, key, expectUUID string) (udid, password, profileUUID string, zipData []byte, err error) {
	if len(key) != 16 || len(blob) <= aes.BlockSize || (len(blob)-aes.BlockSize)%aes.BlockSize != 0 {
		return "", "", "", nil, fmt.Errorf("bad mango")
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", "", "", nil, err
	}
	plain := make([]byte, len(blob)-aes.BlockSize)
	cipher.NewCBCDecrypter(block, blob[:aes.BlockSize]).CryptBlocks(plain, blob[aes.BlockSize:])
	plain, err = pkcs7Unpad(plain)
	if err != nil {
		return "", "", "", nil, err
	}
	if len(plain) <= aes.BlockSize*7 {
		return "", "", "", nil, fmt.Errorf("short mango plain")
	}
	udid = strings.TrimSpace(string(plain[:aes.BlockSize*3]))
	password = strings.TrimSpace(string(plain[aes.BlockSize*3 : aes.BlockSize*4]))
	profileUUID = strings.TrimSpace(string(plain[aes.BlockSize*4 : aes.BlockSize*7]))
	zipData = plain[aes.BlockSize*7:]
	if udid == "" || password == "" || len(zipData) == 0 {
		return "", "", "", nil, fmt.Errorf("empty mango")
	}
	if expectUUID != "" && !strings.EqualFold(profileUUID, expectUUID) {
		return "", "", "", nil, fmt.Errorf("profile mismatch")
	}
	return udid, password, profileUUID, zipData, nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	if padding == 0 {
		padding = blockSize
	}
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty")
	}
	n := int(data[len(data)-1])
	if n == 0 || n > aes.BlockSize || n > len(data) {
		return nil, fmt.Errorf("bad pkcs7")
	}
	for i := 0; i < n; i++ {
		if data[len(data)-1-i] != byte(n) {
			return nil, fmt.Errorf("bad pkcs7")
		}
	}
	return data[:len(data)-n], nil
}
