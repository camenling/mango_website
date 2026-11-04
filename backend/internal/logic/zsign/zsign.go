package zsign

/*
#cgo LDFLAGS: -L/usr/local/lib -lzsign
#cgo CFLAGS: -I/usr/local/include/zsign

#include "zsign.h"
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// SignParam 签名参数（对应 libzsign ZSignOptions）
type SignParam struct {
	IpaPath          string
	P12Path          string
	ProvisionPath    string
	OutputPath       string
	CertPassword     string
	BundleID         string
	BundleName       string
	BundleVersion    string
	EntitlementsPath string
	TempFolder       string
	ZipLevel         int
	Dylibs           []string
}

func Sign(ctx context.Context, params SignParam) (g.Map, error) {
	var signResult C.int
	err := g.Try(ctx, func(ctx context.Context) {
		if params.IpaPath == "" || params.P12Path == "" || params.ProvisionPath == "" || params.CertPassword == "" || params.OutputPath == "" {
			panic(gerror.New("缺乏必要签名参数"))
		}

		if params.ZipLevel <= 0 || params.ZipLevel > 9 {
			params.ZipLevel = 1
		}

		if params.TempFolder == "" {
			tempDir := g.Cfg().MustGet(ctx, "sign.signTempDir").String()
			if tempDir == "" {
				publicPath := g.Cfg().MustGet(ctx, "server.serverRoot").String()
				tempDir = filepath.Join(publicPath, "signTemp")
			}
			if err := os.MkdirAll(tempDir, 0o755); err != nil {
				panic(gerror.Wrap(err, "创建临时文件夹失败"))
			}
			params.TempFolder = tempDir
		}

		var dylibPtrs []*C.char
		for _, dylibPath := range params.Dylibs {
			dylibPtrs = append(dylibPtrs, C.CString(dylibPath))
		}

		opts := C.ZSignOptions{
			ipa_path:       stringToCChar(params.IpaPath),
			p12_path:       stringToCChar(params.P12Path),
			provision_path: stringToCChar(params.ProvisionPath),
			password:       stringToCChar(params.CertPassword),
			output_path:    stringToCChar(params.OutputPath),
			bundle_id:      stringToCChar(params.BundleID),
			bundle_name:    stringToCChar(params.BundleName),
			bundle_version: stringToCChar(params.BundleVersion),
			entitlements:   stringToCChar(params.EntitlementsPath),
			temp_folder:    stringToCChar(params.TempFolder),
			zip_level:      C.int(params.ZipLevel),
			force:          boolToCInt(true),
			dylib_paths:    nil,
		}

		if len(dylibPtrs) > 0 {
			opts.dylib_paths = (**C.char)(unsafe.Pointer(&dylibPtrs[0]))
			opts.dylib_count = C.int(len(dylibPtrs))
		}

		defer func() {
			freeCString(opts.ipa_path)
			freeCString(opts.p12_path)
			freeCString(opts.provision_path)
			freeCString(opts.output_path)
			freeCString(opts.password)
			freeCString(opts.bundle_id)
			freeCString(opts.bundle_name)
			freeCString(opts.bundle_version)
			freeCString(opts.entitlements)
			freeCString(opts.temp_folder)
			for _, ptr := range dylibPtrs {
				freeCString(ptr)
			}
		}()

		signResult = C.ZSign(&opts)
	})

	if err != nil {
		return nil, err
	}

	return handleResult(signResult), nil
}

func boolToCInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

func freeCString(ptr *C.char) {
	if ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

func stringToCChar(s string) *C.char {
	if s == "" {
		return nil
	}
	return C.CString(s)
}

func handleResult(ret C.int) g.Map {
	switch ret {
	case 0:
		return g.Map{"code": 0, "message": "签名成功"}
	case 1:
		return g.Map{"code": 1, "message": "zsign 帮助信息"}
	case 2:
		return g.Map{"code": 2, "message": "zsign 版本信息"}
	default:
		return g.Map{"code": ret, "message": fmt.Sprintf("签名失败，错误码: %d", int(ret))}
	}
}
