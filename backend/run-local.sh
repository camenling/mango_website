#!/usr/bin/env bash
# 本地运行：macOS 用 /usr/local/lib；Linux 裸机用 build/zsign（与 sign/backend 一致）
set -euo pipefail
cd "$(dirname "$0")"

MAIN="${1:-./main}"

case "$(uname -s)" in
Darwin)
	export CGO_ENABLED=1
	export CGO_LDFLAGS="-L/usr/local/lib"
	export DYLD_LIBRARY_PATH="/usr/local/lib:${DYLD_LIBRARY_PATH:-}"
	;;
Linux)
	export LD_LIBRARY_PATH="${PWD}/build/zsign:${LD_LIBRARY_PATH:-}"
	;;
*)
	echo "不支持的操作系统" >&2
	exit 1
	;;
esac

if [[ ! -x "${MAIN}" ]] && [[ ! -f "${MAIN}" ]]; then
	MAIN="./build/go/main"
fi

if [[ ! -f "${MAIN}" ]]; then
	echo "找不到可执行文件，请先 go build 或 make go-builder" >&2
	exit 1
fi

exec "${MAIN}"
