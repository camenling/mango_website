#!/usr/bin/env bash
# macOS 本机：编译 zsign 并安装到 /usr/local（与 sign/backend 一致；不写入 build/zsign）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${ROOT}/manifest/docker/zsign/build/macos"
make clean && make INSTALL_NAME=/usr/local/lib/libzsign.so all
sudo mkdir -p /usr/local/lib /usr/local/include
sudo cp ../../bin/libzsign.so /usr/local/lib/
sudo chmod 755 /usr/local/lib/libzsign.so
sudo install_name_tool -id /usr/local/lib/libzsign.so /usr/local/lib/libzsign.so
sudo rm -rf /usr/local/include/zsign
sudo cp -r ../../src /usr/local/include/zsign
echo "已安装: /usr/local/lib/libzsign.so、/usr/local/include/zsign"
