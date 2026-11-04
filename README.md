# iOS 签名工具

仿 [signapp.cn](https://signapp.cn/) / [sign.lc](https://www.sign.lc/) 的 iOS 签名工具宣传首页 + P12 证书在线签名安装器。

## 功能

- **宣传首页**：产品介绍、功能特性、安装流程说明
- **证书安装页**：上传 P12 + mobileprovision + 密码，服务器签名后 OTA 安装
- **证书查询**：解析证书有效期、Team、Bundle ID 等信息
- **后端**：GoFrame v2，支持 zsign 签名引擎
- **前端**：Vue 3 + Ant Design Vue + Vite

## 项目结构

```
mgqqianduan/
├── backend/          # GoFrame 后端
│   ├── main.go
│   ├── manifest/config/config.yaml
│   └── internal/
├── frontend/         # Vue 3 前端
│   └── src/
│       ├── views/Home.vue      # 宣传首页
│       └── views/Install.vue   # P12 证书安装页
└── README.md
```

## 快速开始

### 1. 后端

```bash
cd backend
go mod tidy
go run main.go
```

服务默认运行在 `http://localhost:8199`

### 2. 前端开发

```bash
cd frontend
npm install
npm run dev
```

开发服务器 `http://localhost:5173`，API 代理到后端。

### 3. 生产构建

```bash
cd frontend
npm run build
# 产物输出到 backend/resource/public/

cd ../backend
go run main.go
```

## 部署配置

编辑 `backend/manifest/config/config.yaml`：

| 配置项 | 说明 |
|--------|------|
| `sign.authBaseUrl` | 授权后台地址，用于取最新已发布 IPA |
| `sign.authAesKey` | 与授权 `siteSign.aesKey` 一致，请求/响应 AES 加密 |
| `sign.baseIpa` | 无已发布版本时的兜底 IPA 路径 |
| `sign.certMangoKey` | 内置 `cert.mango` 的 16 字节 AES 钥匙，须与授权 `deviceAuth.certMangoKey` 一致 |
| `sign.siteUrl` | 对外访问域名（HTTPS），用于生成 OTA 安装链接 |
| `sign.zsignPath` | zsign 可执行文件路径 |
| `sign.appName` | 安装时显示的应用名称 |

### 签名底包

官网会向授权后台 `POST /api/v1/sign/app/latest-published`（AES 加密，且出口 IP 须在授权 `siteSign.allowIps`）取最新已发布记录，再 `POST /latest-published/file` 下载并缓存到 `sign.ipaCacheDir`。

仅当没有已发布版本时，才使用 `backend/resource/base/base.ipa`。

### 安装 zsign（Linux/macOS 服务器）

```bash
git clone https://github.com/zhlynn/zsign.git
cd zsign && g++ -o zsign *.cpp -lcrypto -O3
# 将 zsign 路径写入 config.yaml 的 sign.zsignPath
```

> iOS OTA 安装要求 HTTPS，生产环境请配置 SSL 证书。

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/sign/sign` | 上传证书并签名（multipart） |
| POST | `/api/sign/check-cert` | 查询证书状态 |
| GET | `/api/sign/health` | 健康检查 |

### 签名请求参数

- `p12`：P12 证书文件
- `mobileprovision`：描述文件
- `password`：证书密码（仅字母数字）

### 签名响应

```json
{
  "code": 0,
  "data": {
    "taskId": "...",
    "installUrl": "itms-services://?action=download-manifest&url=...",
    "plistUrl": "https://your.domain/signed/xxx/manifest.plist",
    "ipaUrl": "https://your.domain/signed/xxx/app.ipa",
    "message": "签名完成"
  }
}
```

## 页面路由

| 路径 | 页面 |
|------|------|
| `/` | 宣传首页 |
| `/install` | P12 证书上传安装 |

## 注意事项

1. 证书材料签名完成后自动删除（可配置 `sign.cleanupCerts`）
2. 密码仅允许字母与数字，防止命令注入
3. 请使用 Safari 浏览器进行 OTA 安装
4. 企业证书可能出现掉签，需及时更新

## License

MIT