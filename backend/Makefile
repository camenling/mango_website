.PHONY: all clean clean-cache docker-prod zsign-builder zsign-builder-mac go-builder build-prod-image pack-prod-deploy \
	docker-pull-base docker-save-base docker-load-base docker-base-ensure

BUILD_DIR := ./build
DOCKER_CACHE_DIR := ${BUILD_DIR}/docker-cache
BASE_IMAGES_DIR := ./manifest/docker/base-images
DOCKER_BASE_SCRIPT := ./scripts/docker-base-images.sh
ZSIGN_LIB := ${BUILD_DIR}/zsign/libzsign.so
GO_BIN := ${BUILD_DIR}/go/main

export DOCKER_BUILDKIT=1
export BUILDX_BUILDER ?= desktop-linux
BASE_IMAGE_GO ?= golang:1.23-alpine
BASE_IMAGE_ALPINE ?= alpine:3.18
DOCKER_BUILD_FLAGS = --pull=false \
	--build-arg BASE_IMAGE=$(BASE_IMAGE_GO) \
	--cache-from type=local,src=$(DOCKER_CACHE_DIR) \
	--cache-to type=local,dest=$(DOCKER_CACHE_DIR),mode=max

DOCKER_BUILD_FLAGS_ZSIGN = --pull=false \
	--build-arg BASE_IMAGE=$(BASE_IMAGE_ALPINE) \
	--cache-from type=local,src=$(DOCKER_CACHE_DIR) \
	--cache-to type=local,dest=$(DOCKER_CACHE_DIR),mode=max

all: zsign-builder go-builder docker-prod

build-prod-image:
	@chmod +x ./build-prod-image.sh
	@./build-prod-image.sh

pack-prod-deploy:
	@chmod +x ./pack-prod-deploy.sh
	@./pack-prod-deploy.sh --build

clean:
	rm -rf ${BUILD_DIR}

clean-cache:
	docker builder prune -f
	rm -rf $(DOCKER_CACHE_DIR)

docker-base-ensure:
	@chmod +x $(DOCKER_BASE_SCRIPT)
	@$(DOCKER_BASE_SCRIPT) ensure

docker-pull-base: docker-base-ensure

docker-save-base:
	@chmod +x $(DOCKER_BASE_SCRIPT)
	@$(DOCKER_BASE_SCRIPT) save

docker-load-base:
	@chmod +x $(DOCKER_BASE_SCRIPT)
	@$(DOCKER_BASE_SCRIPT) load

docker-prod: docker-base-ensure
	@test -f ${GO_BIN} || (echo "错误: 缺少 ${GO_BIN}，请先 make go-builder" && exit 1)
	@chmod +x ./scripts/ensure-zsign-linux.sh
	@./scripts/ensure-zsign-linux.sh ${ZSIGN_LIB}
	@mkdir -p ${BUILD_DIR}/docker $(DOCKER_CACHE_DIR)
	docker buildx build --platform linux/amd64 \
		$(DOCKER_BUILD_FLAGS_ZSIGN) \
		--load \
		-f ./manifest/docker/Dockerfile-runtime \
		-t mangosign:prod .
	docker save -o ${BUILD_DIR}/docker/mangosign_prod.tar mangosign:prod

zsign-builder: docker-base-ensure ${ZSIGN_LIB}
${ZSIGN_LIB}:
	@echo "Building zsign..."
	@mkdir -p ${BUILD_DIR}/zsign $(DOCKER_CACHE_DIR)
	docker buildx build --target export \
		$(DOCKER_BUILD_FLAGS_ZSIGN) \
		--platform linux/amd64 \
		-o type=local,dest=${BUILD_DIR}/zsign/ \
		-f ./manifest/docker/Dockerfile-zsign .
	@tar -xzf ${BUILD_DIR}/zsign/zsign-output.tar.gz -C ${BUILD_DIR}/zsign
	@rm ${BUILD_DIR}/zsign/zsign-output.tar.gz

# macOS 本机：安装 libzsign 到 /usr/local（build/zsign 仅 Linux Docker 产物）
zsign-builder-mac:
	@chmod +x ./manifest/docker/build-zsign-mac.sh
	@./manifest/docker/build-zsign-mac.sh

go-builder: docker-base-ensure zsign-linux-ensure ${GO_BIN}

zsign-linux-ensure:
	@chmod +x ./scripts/ensure-zsign-linux.sh
	@./scripts/ensure-zsign-linux.sh ${ZSIGN_LIB} || (rm -f ${ZSIGN_LIB} && $(MAKE) zsign-builder)

${GO_BIN}: zsign-linux-ensure
	@echo "Building Go binary..."
	@mkdir -p ${BUILD_DIR}/go ${BUILD_DIR}/docker $(DOCKER_CACHE_DIR)
	docker buildx build --target export \
		$(DOCKER_BUILD_FLAGS) \
		--platform linux/amd64 \
		-o type=local,dest=${BUILD_DIR}/go \
		-f ./manifest/docker/Dockerfile-go .
