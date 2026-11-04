#!/usr/bin/env bash
# 首次上线 / 发版：打部署 zip（镜像 tar + compose + /app 目录骨架）
# 用法（在 backend/ 目录）:
#   ./pack-prod-deploy.sh              # 要求已有 build/docker/mangosign_prod.tar
#   ./pack-prod-deploy.sh --build      # 先 ./build-prod-image.sh 再打包
#   ./pack-prod-deploy.sh --build --full
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"

IMAGE_TAR="build/docker/mangosign_prod.tar"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
STAGE="${SCRIPT_DIR}/build/deploy-stage"
OUTPUT_ZIP="${SCRIPT_DIR}/build/deploy/mangosign_deploy_${TIMESTAMP}.zip"

DO_BUILD=0
FULL=0

while [[ $# -gt 0 ]]; do
	case "$1" in
	--build) DO_BUILD=1 ;;
	--full) FULL=1; DO_BUILD=1 ;;
	-h | --help)
		cat <<'EOF'
pack-prod-deploy.sh — 生成服务器首次上线 zip

  ./pack-prod-deploy.sh           仅打包（需已有 build/docker/mangosign_prod.tar）
  ./pack-prod-deploy.sh --build     先 build-prod-image.sh 再打包
  ./pack-prod-deploy.sh --build --full  全量编译 zsign + Go + 镜像后打包

zip 内含（解压目录即 compose 工作目录，整体挂载为容器 /app）:
  mangosign_prod.tar, docker-compose.yaml, .env, README.md
  manifest/config/config.yaml
  resource/base/base.ipa, resource/public/（前端）, resource/temp/
  resource/public/signed, resource/public/signTemp（空）, redis-data/
EOF
		exit 0
		;;
	*)
		echo "未知参数: $1" >&2
		exit 1
		;;
	esac
	shift
done

if [[ "${DO_BUILD}" -eq 1 ]]; then
	if [[ "${FULL}" -eq 1 ]]; then
		./build-prod-image.sh --full
	else
		./build-prod-image.sh
	fi
fi

if [[ ! -f "${IMAGE_TAR}" ]]; then
	echo "缺少 ${IMAGE_TAR}，请先执行 ./build-prod-image.sh 或 ./pack-prod-deploy.sh --build" >&2
	exit 1
fi

rm -rf "${STAGE}"
mkdir -p "${STAGE}/manifest/config"
mkdir -p "${STAGE}/resource/base" "${STAGE}/resource/temp"
mkdir -p "${STAGE}/resource/public/signed" "${STAGE}/resource/public/signTemp" "${STAGE}/resource/public/themes"
mkdir -p "${STAGE}/resource/themes"
mkdir -p "${STAGE}/redis-data"
mkdir -p "${SCRIPT_DIR}/build/deploy"

cp -f "${IMAGE_TAR}" "${STAGE}/mangosign_prod.tar"
cp -f docker-compose.yaml "${STAGE}/docker-compose.yaml"
cp -f deploy/config.yaml.example "${STAGE}/manifest/config/config.yaml"
cp -f .env.example "${STAGE}/.env"
cp -f deploy/README.md "${STAGE}/README.md"

if [[ -f resource/base/base.ipa ]]; then
	cp -f resource/base/base.ipa "${STAGE}/resource/base/base.ipa"
else
	echo "警告: 未找到 resource/base/base.ipa，包内仅有空目录，上线前请放入 IPA" >&2
	touch "${STAGE}/resource/base/.keep"
fi

if [[ -f resource/public/index.html ]]; then
	if command -v rsync >/dev/null 2>&1; then
		rsync -a --exclude 'signed/' --exclude 'signTemp/' resource/public/ "${STAGE}/resource/public/"
	else
		echo "警告: 未安装 rsync，跳过 resource/public 静态文件；请本机 npm run build 后重打 zip" >&2
	fi
else
	echo "警告: 未找到 resource/public/index.html，请先 cd frontend && npm run build" >&2
fi
mkdir -p "${STAGE}/resource/public/signed" "${STAGE}/resource/public/signTemp"

if [[ -d resource/themes ]]; then
	if command -v rsync >/dev/null 2>&1; then
		rsync -a resource/themes/ "${STAGE}/resource/themes/"
	fi
fi

(
	cd "${STAGE}"
	zip -r -q "${OUTPUT_ZIP}" .
)

rm -rf "${STAGE}"

echo ""
echo "✅ 部署包: ${OUTPUT_ZIP}"
echo ""
echo "上传到服务器后:"
echo "  unzip mangosign_deploy_*.zip -d /opt/mangosign && cd /opt/mangosign"
echo "  vi manifest/config/config.yaml    # 修改 sign.siteUrl"
echo "  docker load -i mangosign_prod.tar"
echo "  docker compose up -d"
