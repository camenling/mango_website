package controller

import (
	"encoding/base64"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"

	"mangosign/internal/pkg/certzip"
)

type resolvedCerts struct {
	p12Path       string
	provisionPath string
	p12Name       string
	provisionName string
	fromZip       bool
}

// CertBase64Payload 与主站 DeviceUploadCheckReq 一致：p12/mp 为文件内容的 Base64
type CertBase64Payload struct {
	P12             string `json:"p12" form:"p12"`
	Mp              string `json:"mp" form:"mp"`
	Mobileprovision string `json:"mobileprovision" form:"mobileprovision"`
}

func resolveCertUploads(r *ghttp.Request, taskDir string, base64 CertBase64Payload) (*resolvedCerts, error) {
	p12B64 := strings.TrimSpace(base64.P12)
	mpB64 := strings.TrimSpace(base64.Mp)
	if mpB64 == "" {
		mpB64 = strings.TrimSpace(base64.Mobileprovision)
	}
	if p12B64 != "" && mpB64 != "" {
		return writeResolvedCertsFromBase64(p12B64, mpB64, taskDir, "cert.p12", "profile.mobileprovision")
	}

	if certs, ok, err := tryResolveCertBase64(r, taskDir); ok || err != nil {
		return certs, err
	}

	zipFile := r.GetUploadFile("certZip")
	p12File := r.GetUploadFile("p12")
	provisionFile := r.GetUploadFile("mobileprovision")

	if zipFile != nil {
		zipPath := gfile.Join(taskDir, "certs.zip")
		if err := saveUpload(zipFile, zipPath); err != nil {
			return nil, gerror.Wrap(err, "保存 ZIP 文件失败")
		}
		extractDir := gfile.Join(taskDir, "extracted")
		files, err := certzip.ExtractZip(zipPath, extractDir)
		if err != nil {
			return nil, err
		}
		return &resolvedCerts{
			p12Path:       files.P12Path,
			provisionPath: files.MobileprovisionPath,
			p12Name:       files.P12Name,
			provisionName: files.ProvisionName,
			fromZip:       true,
		}, nil
	}

	if p12File == nil {
		return nil, gerror.New("请上传 P12 证书文件，或上传包含证书的 ZIP 压缩包")
	}
	if provisionFile == nil {
		return nil, gerror.New("请上传描述文件 (.mobileprovision)，或上传包含证书的 ZIP 压缩包")
	}

	p12Path := gfile.Join(taskDir, "cert.p12")
	provisionPath := gfile.Join(taskDir, "profile.mobileprovision")
	if err := saveUpload(p12File, p12Path); err != nil {
		return nil, gerror.Wrap(err, "保存 P12 文件失败")
	}
	if err := saveUpload(provisionFile, provisionPath); err != nil {
		return nil, gerror.Wrap(err, "保存描述文件失败")
	}

	return &resolvedCerts{
		p12Path:       p12Path,
		provisionPath: provisionPath,
		p12Name:       p12File.Filename,
		provisionName: provisionFile.Filename,
	}, nil
}

func tryResolveCertBase64(r *ghttp.Request, taskDir string) (*resolvedCerts, bool, error) {
	p12B64 := strings.TrimSpace(r.Get("p12").String())
	mpB64 := strings.TrimSpace(r.Get("mp").String())
	if mpB64 == "" {
		mpB64 = strings.TrimSpace(r.Get("mobileprovision").String())
	}

	if p12B64 == "" || mpB64 == "" {
		var body CertBase64Payload
		if err := r.Parse(&body); err == nil {
			if p12B64 == "" {
				p12B64 = strings.TrimSpace(body.P12)
			}
			if mpB64 == "" {
				mpB64 = strings.TrimSpace(body.Mp)
				if mpB64 == "" {
					mpB64 = strings.TrimSpace(body.Mobileprovision)
				}
			}
		}
	}

	if p12B64 == "" || mpB64 == "" {
		return nil, false, nil
	}

	certs, err := writeResolvedCertsFromBase64(p12B64, mpB64, taskDir, "cert.p12", "profile.mobileprovision")
	if err != nil {
		return nil, true, err
	}
	return certs, true, nil
}

func writeResolvedCertsFromBase64(p12B64, mpB64, taskDir, p12Name, provisionName string) (*resolvedCerts, error) {
	p12Path := gfile.Join(taskDir, "cert.p12")
	provisionPath := gfile.Join(taskDir, "profile.mobileprovision")
	if err := writeBase64File(p12B64, p12Path); err != nil {
		return nil, gerror.Wrap(err, "保存 P12 文件失败")
	}
	if err := writeBase64File(mpB64, provisionPath); err != nil {
		return nil, gerror.Wrap(err, "保存描述文件失败")
	}
	if p12Name == "" {
		p12Name = "cert.p12"
	}
	if provisionName == "" {
		provisionName = "profile.mobileprovision"
	}
	return &resolvedCerts{
		p12Path:       p12Path,
		provisionPath: provisionPath,
		p12Name:       p12Name,
		provisionName: provisionName,
	}, nil
}

func writeBase64File(raw, path string) error {
	raw = strings.TrimSpace(raw)
	if i := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && i >= 0 {
		raw = raw[i+1:]
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return gerror.New("证书 Base64 格式无效")
	}
	if len(data) == 0 {
		return gerror.New("证书内容为空")
	}
	return gfile.PutBytes(path, data)
}

func isZipFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".zip")
}
