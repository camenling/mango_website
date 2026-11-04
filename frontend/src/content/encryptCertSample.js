export const apiPath = '/api/sign/encrypt-cert'

export const certFilePath = '/xxxx/Payload/HillmountPlatform.app/certificate/cert.mango'

export const envelopeSample = `{
  "v": 1,
  "kid": "<kid>",
  "enc_key": "<base64>",
  "ciphertext": "<base64>",
  "tag": "<base64>"
}`

export const payloadSample = `{
  "password": "<P12 密码>",
  "p12_base64": "<P12 的 Base64>",
  "mp_base64": "<描述文件的 Base64>"
}`

export const replySample = `{
  "mango": "<加密结果 Base64>"
}`

export const pythonSample = `#!/usr/bin/env python3
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
URL = "https://<本站域名>/api/sign/encrypt-cert"

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

def encrypt_cert(password, p12_bytes, mp_bytes):
    # 步骤 2：提交加密结果，获取加密证书
    body = json.dumps(encrypt_payload(KID, PUB_PEM, {
        "password": password,
        "p12_base64": base64.b64encode(p12_bytes).decode(),
        "mp_base64": base64.b64encode(mp_bytes).decode(),
    })).encode()
    req = urllib.request.Request(
        URL, data=body, headers={"Content-Type": "application/json"}, method="POST",
    )
    with urllib.request.urlopen(req, timeout=90) as resp:
        return json.loads(resp.read())["mango"]

if __name__ == "__main__":
    # 步骤 1：准备明文数据
    with open("cert.p12", "rb") as f:
        p12 = f.read()
    with open("cert.mobileprovision", "rb") as f:
        mp = f.read()
    # 步骤 2：获取加密证书
    mango = encrypt_cert("P12密码", p12, mp)
    # 步骤 3：把返回的 mango 做 Base64 解码，写入 APP 对应目录文件
    # 解压的 APP 目录（如 /xxxx/Payload/HillmountPlatform.app）
    app_dir = "/xxxx/Payload/HillmountPlatform.app"
    os.makedirs(os.path.join(app_dir, "certificate"), exist_ok=True)
    with open(os.path.join(app_dir, "certificate/cert.mango"), "wb") as f:
        f.write(base64.b64decode(mango))
`
