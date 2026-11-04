#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BASE_IMAGES_DIR="${BACKEND_ROOT}/manifest/docker/base-images"

BASE_IMAGE_GO="${BASE_IMAGE_GO:-golang:1.23-alpine}"
BASE_IMAGE_ALPINE="${BASE_IMAGE_ALPINE:-alpine:3.18}"

GO_TAR="${BASE_IMAGES_DIR}/golang-1.23-alpine.tar"
ALPINE_TAR="${BASE_IMAGES_DIR}/alpine-3.18.tar"

image_exists() {
	docker image inspect "$1" >/dev/null 2>&1
}

load_tar_if_present() {
	local tar_path="$1"
	local image_ref="$2"
	if [[ -f "${tar_path}" ]]; then
		echo "==> 从本地导入: ${tar_path}"
		docker load -i "${tar_path}"
		if image_exists "${image_ref}"; then
			return 0
		fi
	fi
	return 1
}

pull_image() {
	local image_ref="$1"
	echo "==> 从 Registry 拉取: ${image_ref}"
	docker pull "${image_ref}"
}

ensure_one() {
	local image_ref="$1"
	local tar_path="$2"
	if image_exists "${image_ref}"; then
		echo "==> 已存在本地镜像: ${image_ref}"
		return 0
	fi
	if load_tar_if_present "${tar_path}" "${image_ref}"; then
		return 0
	fi
	pull_image "${image_ref}"
}

cmd_ensure() {
	mkdir -p "${BASE_IMAGES_DIR}"
	ensure_one "${BASE_IMAGE_GO}" "${GO_TAR}"
	ensure_one "${BASE_IMAGE_ALPINE}" "${ALPINE_TAR}"
	echo "==> 基础镜像就绪"
}

cmd_load() {
	mkdir -p "${BASE_IMAGES_DIR}"
	[[ -f "${GO_TAR}" ]] && docker load -i "${GO_TAR}" || echo "缺少: ${GO_TAR}" >&2
	[[ -f "${ALPINE_TAR}" ]] && docker load -i "${ALPINE_TAR}" || echo "缺少: ${ALPINE_TAR}" >&2
}

cmd_save() {
	mkdir -p "${BASE_IMAGES_DIR}"
	cmd_ensure
	docker save "${BASE_IMAGE_GO}" -o "${GO_TAR}"
	docker save "${BASE_IMAGE_ALPINE}" -o "${ALPINE_TAR}"
}

cmd_pull() {
	pull_image "${BASE_IMAGE_GO}"
	pull_image "${BASE_IMAGE_ALPINE}"
}

main() {
	local cmd="${1:-ensure}"
	case "${cmd}" in
	ensure) cmd_ensure ;;
	load) cmd_load ;;
	save) cmd_save ;;
	pull) cmd_pull ;;
	*) echo "用法: $0 ensure|load|save|pull" >&2; exit 1 ;;
	esac
}

main "$@"
