export const channelFilePath = '/xxxx/Payload/HillmountPlatform.app/website/hm-ota-channel.json'

export const channelPayloadSample = `{
  "HMBuildChannel": "<渠道标识>"
}`

export const apiPath = '/api/sign/channel-blacklist'

export const unblacklistApiPath = '/api/sign/channel-unblacklist'

export const envelopeSample = `{
  "v": 1,
  "kid": "<kid>",
  "enc_key": "<base64>",
  "ciphertext": "<base64>",
  "tag": "<base64>"
}`

export const payloadSample = `{
  "udid": "<设备 UDID>",
  "reason": "<可选>"
}`

export const unblacklistPayloadSample = `{
  "udid": "<设备 UDID>"
}`

export const replySample = `{
  "ok": true,
  "channel": "<渠道标识>"
}`

export const errorSample = `{
  "error": "<错误码>"
}`

export const channelPythonSample = `#!/usr/bin/env python3
# pip install cryptography
import os, json, base64
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

KID = "你的 kid"
PUB_PEM = b"""-----BEGIN PUBLIC KEY-----
你的公钥
-----END PUBLIC KEY-----
"""

def encrypt_payload(kid, pub_pem, plain_obj):
    pub = serialization.load_pem_public_key(pub_pem)
    plain = json.dumps(plain_obj, separators=(",", ":")).encode()
    aes_key, nonce = os.urandom(32), os.urandom(12)
    sealed = AESGCM(aes_key).encrypt(nonce, plain, f"1|{kid}".encode())
    ciphertext, tag = sealed[:-16], sealed[-16:]
    enc_key = pub.encrypt(
        aes_key + nonce,
        padding.OAEP(
            mgf=padding.MGF1(algorithm=hashes.SHA256()),
            algorithm=hashes.SHA256(),
            label=None,
        ),
    )
    return {
        "v": 1,
        "kid": kid,
        "enc_key": base64.b64encode(enc_key).decode(),
        "ciphertext": base64.b64encode(ciphertext).decode(),
        "tag": base64.b64encode(tag).decode(),
    }

if __name__ == "__main__":
    # 步骤 1：准备明文数据
    plain = {"HMBuildChannel": "<渠道标识>"}
    # 步骤 2：获取加密渠道标识（本地加密，不请求接口）
    env = encrypt_payload(KID, PUB_PEM, plain)
    # 步骤 3：写入加密 JSON
    # 解压的 APP 目录（如 /xxxx/Payload/HillmountPlatform.app）
    app_dir = "/xxxx/Payload/HillmountPlatform.app"
    os.makedirs(os.path.join(app_dir, "website"), exist_ok=True)
    with open(os.path.join(app_dir, "website/hm-ota-channel.json"), "w", encoding="utf-8") as f:
        json.dump(env, f, separators=(",", ":"))
`

export const pythonSample = `#!/usr/bin/env python3
# pip install cryptography
import os, json, base64, urllib.request, urllib.error
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

KID = "你的 kid"
PUB_PEM = b"""-----BEGIN PUBLIC KEY-----
你的公钥
-----END PUBLIC KEY-----
"""
URL = "https://<本站域名>/api/sign/channel-blacklist"

def encrypt_payload(kid, pub_pem, plain_obj):
    pub = serialization.load_pem_public_key(pub_pem)
    plain = json.dumps(plain_obj, separators=(",", ":")).encode()
    aes_key, nonce = os.urandom(32), os.urandom(12)
    sealed = AESGCM(aes_key).encrypt(nonce, plain, f"1|{kid}".encode())
    ciphertext, tag = sealed[:-16], sealed[-16:]
    enc_key = pub.encrypt(
        aes_key + nonce,
        padding.OAEP(
            mgf=padding.MGF1(algorithm=hashes.SHA256()),
            algorithm=hashes.SHA256(),
            label=None,
        ),
    )
    return {
        "v": 1,
        "kid": kid,
        "enc_key": base64.b64encode(enc_key).decode(),
        "ciphertext": base64.b64encode(ciphertext).decode(),
        "tag": base64.b64encode(tag).decode(),
    }

def add_blacklist(udid, reason=""):
    # 步骤 1：准备明文数据
    plain = {"udid": udid}
    if reason:
        plain["reason"] = reason
    body = json.dumps(encrypt_payload(KID, PUB_PEM, plain)).encode()
    # 步骤 2：加密请求 POST https://<本站域名>/api/sign/channel-blacklist
    req = urllib.request.Request(
        URL, data=body, headers={"Content-Type": "application/json"}, method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=90) as resp:
            status = resp.status
            raw = resp.read()
    except urllib.error.HTTPError as e:
        status = e.code
        raw = e.read()
    # 步骤 3：解析返回
    return status, json.loads(raw)

if __name__ == "__main__":
    status, data = add_blacklist("00008020-001234567890ABCD")
    print(status, data)
`

export const unblacklistPythonSample = `#!/usr/bin/env python3
# pip install cryptography
import os, json, base64, urllib.request, urllib.error
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

KID = "你的 kid"
PUB_PEM = b"""-----BEGIN PUBLIC KEY-----
你的公钥
-----END PUBLIC KEY-----
"""
URL = "https://<本站域名>/api/sign/channel-unblacklist"

def encrypt_payload(kid, pub_pem, plain_obj):
    pub = serialization.load_pem_public_key(pub_pem)
    plain = json.dumps(plain_obj, separators=(",", ":")).encode()
    aes_key, nonce = os.urandom(32), os.urandom(12)
    sealed = AESGCM(aes_key).encrypt(nonce, plain, f"1|{kid}".encode())
    ciphertext, tag = sealed[:-16], sealed[-16:]
    enc_key = pub.encrypt(
        aes_key + nonce,
        padding.OAEP(
            mgf=padding.MGF1(algorithm=hashes.SHA256()),
            algorithm=hashes.SHA256(),
            label=None,
        ),
    )
    return {
        "v": 1,
        "kid": kid,
        "enc_key": base64.b64encode(enc_key).decode(),
        "ciphertext": base64.b64encode(ciphertext).decode(),
        "tag": base64.b64encode(tag).decode(),
    }

def remove_blacklist(udid):
    # 步骤 1：准备明文数据
    plain = {"udid": udid}
    body = json.dumps(encrypt_payload(KID, PUB_PEM, plain)).encode()
    # 步骤 2：加密请求 POST https://<本站域名>/api/sign/channel-unblacklist
    req = urllib.request.Request(
        URL, data=body, headers={"Content-Type": "application/json"}, method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=90) as resp:
            status = resp.status
            raw = resp.read()
    except urllib.error.HTTPError as e:
        status = e.code
        raw = e.read()
    # 步骤 3：解析返回
    return status, json.loads(raw)

if __name__ == "__main__":
    status, data = remove_blacklist("00008020-001234567890ABCD")
    print(status, data)
`
