#!/usr/bin/env bash
# 日常：改 Go 后重编二进制 + 打运行时镜像
#   ./build-prod-image.sh           # 复用 libzsign，重编 Go + 镜像
#   ./build-prod-image.sh --full    # 强制重编 zsign + Go + 镜像
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"

export BUILDX_BUILDER="${BUILDX_BUILDER:-desktop-linux}"
docker buildx use "${BUILDX_BUILDER}" >/dev/null 2>&1 || true

BUILD_DIR="./build"
GO_BIN="${BUILD_DIR}/go/main"
ZSIGN_LIB="${BUILD_DIR}/zsign/libzsign.so"
IMAGE_TAR="${BUILD_DIR}/docker/mangosign_prod.tar"
RELEASE_DIR="${BUILD_DIR}/docker/releases"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

FULL_BUILD=0
while [[ $# -gt 0 ]]; do
	case "$1" in
	--full) FULL_BUILD=1 ;;
	-h | --help)
		cat <<'EOF'
build-prod-image.sh — 编译 Go + 导出 mangosign:prod 镜像 tar

  ./build-prod-image.sh         复用 Linux libzsign，重编 Go + 镜像
  ./build-prod-image.sh --full  强制 zsign + Go + 镜像

产物:
  build/go/main
  build/docker/mangosign_prod.tar

上线打包 zip（含 compose + 配置）:
  ./pack-prod-deploy.sh --build
EOF
		exit 0
		;;
	*) echo "未知参数: $1（--help 查看）" >&2; exit 1 ;;
	esac
	shift
done

chmod +x ./scripts/docker-base-images.sh
./scripts/docker-base-images.sh ensure

need_zsign=0
if [[ "${FULL_BUILD}" -eq 1 ]] || [[ ! -f "${ZSIGN_LIB}" ]]; then
	need_zsign=1
elif ! file "${ZSIGN_LIB}" | grep -q 'ELF 64-bit LSB'; then
	echo "==> 当前 libzsign.so 非 Linux ELF，重新编译 zsign..."
	rm -f "${ZSIGN_LIB}"
	need_zsign=1
fi
if [[ "${need_zsign}" -eq 1 ]]; then
	make zsign-builder
fi

rm -f "${GO_BIN}"
make go-builder
rm -f "${IMAGE_TAR}"
make docker-prod

mkdir -p "${RELEASE_DIR}"
cp -f "${IMAGE_TAR}" "${RELEASE_DIR}/mangosign_prod_${TIMESTAMP}.tar"

echo ""
echo "✅ 完成"
echo "   二进制:   ${GO_BIN}"
echo "   镜像包:   ${IMAGE_TAR}"
echo "   留存:     ${RELEASE_DIR}/mangosign_prod_${TIMESTAMP}.tar"
echo ""
echo "打首次上线 zip:"
echo "   ./pack-prod-deploy.sh"
echo ""
echo "服务器:"
echo "   unzip 部署包 && cd 目录"
echo "   docker load -i mangosign_prod.tar"
echo "   docker compose up -d"
