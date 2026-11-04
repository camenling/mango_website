package service

import (
	"context"

	"mangosign/internal/model"
)

type ISign interface {
	SignAndInstall(ctx context.Context, in *model.SignInput) (*model.SignResult, error)
	CheckCert(ctx context.Context, in *model.CertCheckInput) (*model.CertInfo, error)
	StartCleanupWorker(ctx context.Context)
}

var localSign ISign

func Sign() ISign {
	if localSign == nil {
		panic("implement not found for interface ISign, forgot register?")
	}
	return localSign
}

func RegisterSign(i ISign) {
	localSign = i
}
