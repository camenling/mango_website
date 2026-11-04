# zsign + Go + Docker 构建说明（对齐 sign/backend）

## 分工（与 sign/backend 一致）

| 场景 | libzsign 位置 | Go CGO |
|------|----------------|--------|
| **Mac 本机** | `/usr/local/lib/libzsign.so` + 头文件 `/usr/local/include/zsign` | `zsign.go` 固定 `#cgo LDFLAGS: -L/usr/local/lib -lzsign` |
| **Docker 编 Go / 线上** | `build/zsign/libzsign.so`（**仅 Linux ELF**） | `Dockerfile-go` 拷贝到容器 `/usr/local/lib` 后 `CGO_LDFLAGS="-L/usr/local/lib -lzsign -lminizip"` |

`build/zsign/` **只存 Docker `make zsign-builder` 产物**，不要在 Mac 上把 Mach-O 拷进去。

---

## 产物路径

| 文件 | 说明 |
|------|------|
| `build/zsign/libzsign.so` | Linux Alpine ELF |
| `build/zsign/include/` | C 头文件 |
| `build/go/main` | Linux amd64 二进制 |
| `build/docker/mangosign_prod.tar` | 镜像 `mangosign:prod` |

---

## Linux 发版

```bash
./build-prod-image.sh          # 复用 ELF libzsign，重编 Go + 镜像
./build-prod-image.sh --full   # 强制 zsign + Go + 镜像
```

若 `build/zsign/libzsign.so` 被误换成 Mach-O，`make go-builder` / `build-prod-image.sh` 会检测并自动 `make zsign-builder`。

---

## macOS 本机

```bash
make zsign-builder-mac   # 安装到 /usr/local，不碰 build/zsign
export CGO_ENABLED=1
export CGO_LDFLAGS="-L/usr/local/lib"
export DYLD_LIBRARY_PATH=/usr/local/lib
go build -o main main.go
./run-local.sh ./main
```

GoLand 环境变量同上（可加 `LD_LIBRARY_PATH=/usr/local/lib` 无妨）。

---

## 运行时镜像

`Dockerfile-runtime` 将 `build/zsign/libzsign.so` 拷入容器 `/usr/local/lib`，入口 `main`，端口 **8199**。
