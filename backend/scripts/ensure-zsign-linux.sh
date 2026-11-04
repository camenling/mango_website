#!/usr/bin/env bash
# 确保 build/zsign/libzsign.so 为 Linux amd64 ELF（Docker 编 Go / 打镜像用）
set -euo pipefail

ZSIGN_LIB="${1:-./build/zsign/libzsign.so}"

if [[ ! -f "${ZSIGN_LIB}" ]]; then
	echo "缺少 ${ZSIGN_LIB}"
	exit 1
fi

if file "${ZSIGN_LIB}" | grep -q 'ELF 64-bit LSB'; then
	exit 0
fi

echo "错误: ${ZSIGN_LIB} 不是 Linux ELF（勿将 Mac 版 lib 放入 build/zsign）。" >&2
echo "请先: rm -f ${ZSIGN_LIB} && make zsign-builder" >&2
exit 1
