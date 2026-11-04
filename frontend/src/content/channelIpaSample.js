export const checkPath = '/api/sign/channel-ipa-check'
export const apiPath = '/api/sign/channel-ipa'

export const checkSample = `{
  "track": "release",
  "channel": "<渠道标识>",
  "build": "260920"
}`

export const payloadSample = `{
  "track": "release",
  "channel": "<渠道标识>",
  "build": "260922"
}`

export const envelopeSample = `{
  "v": 1,
  "kid": "<kid>",
  "enc_key": "<base64>",
  "ciphertext": "<base64>",
  "tag": "<base64>"
}`

export const themePresetSample = `{
  "v": 1,
  "kid": "<kid>",
  "enc_key": "<base64>",
  "ciphertext": "<base64>",
  "tag": "<base64>",
  "theme": {
    "style": 1,
    "accent": "#0A84FF",
    "glass": true
  }
}`

export const themePackageSample = `{
  "v": 1,
  "kid": "<kid>",
  "enc_key": "<base64>",
  "ciphertext": "<base64>",
  "tag": "<base64>",
  "theme": {
    "package": "<.hmtheme 文件的 base64>"
  }
}`

export const checkReplyUpdateSample = `{
  "ok": true,
  "track": "release",
  "channel": "<渠道标识>",
  "build": "260920",
  "is_latest": false,
  "latest": {
    "track": "release",
    "version": "1.2.3",
    "hm_build_date": "260922"
  }
}`

export const checkReplyLatestSample = `{
  "ok": true,
  "track": "release",
  "channel": "<渠道标识>",
  "build": "260922",
  "is_latest": true,
  "latest": {
    "track": "release",
    "version": "1.2.3",
    "hm_build_date": "260922"
  }
}`

export const errorSample = `{
  "error": "<错误码>"
}`

export const checkPythonSample = `#!/usr/bin/env python3
# pip install cryptography
import os, json, base64, urllib.request
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

KID = "你的 kid"
PUB_PEM = b"""-----BEGIN PUBLIC KEY-----
你的公钥
-----END PUBLIC KEY-----
"""
URL = "https://<本站域名>/api/sign/channel-ipa-check"

def encrypt_payload(kid, pub_pem, plain_obj):
    # 步骤 2：使用 kid 和公钥加密明文字段
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

def check_channel_ipa(track, channel, build):
    # 步骤 1：准备明文数据
    plain = {"track": track, "channel": channel, "build": build}
    # 步骤 2：加密请求 POST https://<本站域名>/api/sign/channel-ipa-check
    body = json.dumps(encrypt_payload(KID, PUB_PEM, plain)).encode()
    req = urllib.request.Request(
        URL, data=body, headers={"Content-Type": "application/json"}, method="POST",
    )
    with urllib.request.urlopen(req, timeout=90) as resp:
        # 步骤 3：解析返回
        return resp.status, json.loads(resp.read())

status, data = check_channel_ipa("release", "<渠道标识>", "260920")
print(status, data)
`

export const downloadPythonSample = `#!/usr/bin/env python3
# pip install cryptography
import os, json, base64, urllib.error, urllib.request
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

KID = "你的 kid"
PUB_PEM = b"""-----BEGIN PUBLIC KEY-----
你的公钥
-----END PUBLIC KEY-----
"""
URL = "https://<本站域名>/api/sign/channel-ipa"

def encrypt_payload(kid, pub_pem, plain_obj):
    # 步骤 2：使用 kid 和公钥加密明文字段
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

def download_channel_ipa(track, channel, build, out_file="mango.ipa"):
    # 步骤 1：准备明文数据
    plain = {"track": track, "channel": channel, "build": build}
    # 步骤 2：加密请求 POST https://<本站域名>/api/sign/channel-ipa
    body = json.dumps(encrypt_payload(KID, PUB_PEM, plain)).encode()
    req = urllib.request.Request(
        URL, data=body, headers={"Content-Type": "application/json"}, method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=900) as resp:
            # 步骤 3：HTTP 200 时按二进制写入 IPA
            with open(out_file, "wb") as f:
                f.write(resp.read())
            return resp.status, None
    except urllib.error.HTTPError as e:
        # 非 200 时响应体是 JSON error
        return e.code, json.loads(e.read())

status, err = download_channel_ipa("release", "<渠道标识>", "260922")
print(status, err)
`
