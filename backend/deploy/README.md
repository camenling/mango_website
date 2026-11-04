# 首次上线部署包说明

Compose 将 **当前目录整体挂载为容器 `/app`**（`volumes: ./:/app`）。镜像内仅提供 Go 二进制（`/usr/bin/main`）与 zsign 库；**配置、前端、IPA、运行时数据均在宿主机此目录**。

## 包内文件

| 路径 | 说明 |
|------|------|
| `mangosign_prod.tar` | Docker 镜像，`docker load -i mangosign_prod.tar` |
| `docker-compose.yaml` | 启动编排 |
| `manifest/config/config.yaml` | 生产配置（**必改** `sign.siteUrl`、`sign.authBaseUrl`） |
| `.env` | 宿主机端口等（默认 `MANGOSIGN_PORT=8199`） |
| `resource/base/base.ipa` | 可选兜底 IPA（有已发布版本时不需要） |
| `resource/public/` | 前端静态（`index.html`、`assets/`），发版可整目录替换 |
| `resource/temp/` | 临时文件 |
| `resource/public/signed/` | 签名产物 IPA/plist |
| `resource/public/signTemp/` | 签名过程目录 |
| `redis-data/` | 可选，配合 `docker compose --profile redis up -d` |

## 服务器步骤（Linux + Docker）

```bash
unzip mangosign_deploy_*.zip -d /opt/mangosign
cd /opt/mangosign

vi manifest/config/config.yaml   # sign.siteUrl、sign.authBaseUrl

docker load -i mangosign_prod.tar
docker compose up -d
```

## 更新版本

**仅后端 / zsign：** 上传新 `mangosign_prod.tar` → `docker load` → `docker compose up -d`（宿主机 `manifest/`、`resource/` 保留）。

**仅前端：** 本机 `npm run build` 后，将 `backend/resource/public/`（勿覆盖线上 `signed/`、`signTemp/` 内业务数据）同步到服务器同路径，**无需**重建镜像。

**全量：** 重新 `./pack-prod-deploy.sh --build`，解压时勿覆盖服务器上已有 `resource/public/signed` 与 `resource/temp`（或先备份再合并）。

## 健康检查

```bash
curl -s http://127.0.0.1:8199/api/sign/health
```

对外需 Nginx/Caddy 配置 HTTPS 反代到 8199。
