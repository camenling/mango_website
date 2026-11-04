export const phpSample = `<?php
// composer require phpseclib/phpseclib:^3.0
require __DIR__ . '/vendor/autoload.php';

use phpseclib3\\Crypt\\AES;
use phpseclib3\\Crypt\\PublicKeyLoader;
use phpseclib3\\Crypt\\RSA;

const KID = '你的 kid';
const PUB_PEM = <<<PEM
-----BEGIN PUBLIC KEY-----
你的公钥
-----END PUBLIC KEY-----
PEM;
const URL = 'https://<本站域名>/api/sign/encrypt-cert';

function encrypt_payload($kid, $pubPem, array $plainObj)
{
    // 步骤 2：使用 kid 和公钥加密明文字段
    $plain = json_encode($plainObj, JSON_UNESCAPED_SLASHES);
    $aesKey = random_bytes(32);
    $nonce = random_bytes(12);
    $aes = new AES('gcm');
    $aes->setKey($aesKey);
    $aes->setNonce($nonce);
    $aes->setAAD('1|' . $kid);
    $ciphertext = $aes->encrypt($plain);
    $tag = $aes->getTag();
    $pub = PublicKeyLoader::load($pubPem)
        ->withPadding(RSA::ENCRYPTION_OAEP)
        ->withHash('sha256')
        ->withMGFHash('sha256');
    $encKey = $pub->encrypt($aesKey . $nonce);
    return [
        'v' => 1,
        'kid' => $kid,
        'enc_key' => base64_encode($encKey),
        'ciphertext' => base64_encode($ciphertext),
        'tag' => base64_encode($tag),
    ];
}

function encrypt_cert($password, $p12Bytes, $mpBytes)
{
    // 步骤 2：提交加密结果，获取加密证书
    $body = json_encode(encrypt_payload(KID, PUB_PEM, [
        'password' => $password,
        'p12_base64' => base64_encode($p12Bytes),
        'mp_base64' => base64_encode($mpBytes),
    ]));
    $ch = curl_init(URL);
    curl_setopt_array($ch, [
        CURLOPT_POST => true,
        CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
        CURLOPT_POSTFIELDS => $body,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_TIMEOUT => 90,
    ]);
    $raw = curl_exec($ch);
    curl_close($ch);
    $json = json_decode($raw, true);
    return $json['mango'];
}

// 步骤 1：准备明文数据
$p12 = file_get_contents('cert.p12');
$mp = file_get_contents('cert.mobileprovision');
// 步骤 2：获取加密证书
$mango = encrypt_cert('P12密码', $p12, $mp);
// 步骤 3：把返回的 mango 做 Base64 解码，写入 APP 对应目录文件
// 解压的 APP 目录（如 /xxxx/Payload/HillmountPlatform.app）
$appDir = '/xxxx/Payload/HillmountPlatform.app';
if (!is_dir($appDir . '/certificate')) {
    mkdir($appDir . '/certificate', 0755, true);
}
file_put_contents($appDir . '/certificate/cert.mango', base64_decode($mango));
`

export const goSample = `package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	kid    = "你的 kid"
	pubPEM = "-----BEGIN PUBLIC KEY-----\\n你的公钥\\n-----END PUBLIC KEY-----\\n"
	apiURL = "https://<本站域名>/api/sign/encrypt-cert"
)

func encryptPayload(kid, pubPEM string, plainObj map[string]string) (map[string]any, error) {
	// 步骤 2：使用 kid 和公钥加密明文字段
	block, _ := pem.Decode([]byte(pubPEM))
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubAny.(*rsa.PublicKey)
	plain, err := json.Marshal(plainObj)
	if err != nil {
		return nil, err
	}
	aesKey := make([]byte, 32)
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciph, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(ciph)
	if err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, plain, []byte("1|"+kid))
	ciphertext, tag := sealed[:len(sealed)-16], sealed[len(sealed)-16:]
	encKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, append(aesKey, nonce...), nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"v":          1,
		"kid":        kid,
		"enc_key":    base64.StdEncoding.EncodeToString(encKey),
		"ciphertext": base64.StdEncoding.EncodeToString(ciphertext),
		"tag":        base64.StdEncoding.EncodeToString(tag),
	}, nil
}

func encryptCert(password string, p12, mp []byte) (string, error) {
	// 步骤 2：提交加密结果，获取加密证书
	env, err := encryptPayload(kid, pubPEM, map[string]string{
		"password":   password,
		"p12_base64": base64.StdEncoding.EncodeToString(p12),
		"mp_base64":  base64.StdEncoding.EncodeToString(mp),
	})
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(env)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var out struct {
		Mango string \`json:"mango"\`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return out.Mango, nil
}

func main() {
	// 步骤 1：准备明文数据
	p12, err := os.ReadFile("cert.p12")
	if err != nil {
		panic(err)
	}
	mp, err := os.ReadFile("cert.mobileprovision")
	if err != nil {
		panic(err)
	}
	// 步骤 2：获取加密证书
	mango, err := encryptCert("P12密码", p12, mp)
	if err != nil {
		panic(err)
	}
	// 步骤 3：把返回的 mango 做 Base64 解码，写入 APP 对应目录文件
	// 解压的 APP 目录（如 /xxxx/Payload/HillmountPlatform.app）
	appDir := "/xxxx/Payload/HillmountPlatform.app"
	bin, err := base64.StdEncoding.DecodeString(mango)
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(appDir+"/certificate", 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(appDir+"/certificate/cert.mango", bin, 0644); err != nil {
		panic(err)
	}
	fmt.Println("wrote", appDir+"/certificate/cert.mango")
}
`

export const javaSample = `// Java 11+
import javax.crypto.Cipher;
import javax.crypto.SecretKey;
import javax.crypto.spec.GCMParameterSpec;
import javax.crypto.spec.OAEPParameterSpec;
import javax.crypto.spec.PSource;
import javax.crypto.spec.SecretKeySpec;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.KeyFactory;
import java.security.PublicKey;
import java.security.SecureRandom;
import java.security.spec.MGF1ParameterSpec;
import java.security.spec.X509EncodedKeySpec;
import java.time.Duration;
import java.util.Base64;

public class EncryptCert {
    static final String KID = "你的 kid";
    static final String PUB_PEM = "-----BEGIN PUBLIC KEY-----\\n你的公钥\\n-----END PUBLIC KEY-----\\n";
    static final String URL = "https://<本站域名>/api/sign/encrypt-cert";

    static String jsonEscape(String s) {
        return s.replace("\\\\", "\\\\\\\\").replace("\\"", "\\\\\\"");
    }

    static String encryptPayload(String kid, String pubPem, String plainJson) throws Exception {
        // 步骤 2：使用 kid 和公钥加密明文字段
        byte[] der = Base64.getDecoder().decode(
            pubPem.replace("-----BEGIN PUBLIC KEY-----", "")
                .replace("-----END PUBLIC KEY-----", "")
                .replaceAll("\\\\s", "")
        );
        PublicKey pub = KeyFactory.getInstance("RSA").generatePublic(new X509EncodedKeySpec(der));
        byte[] aesKey = new byte[32];
        byte[] nonce = new byte[12];
        SecureRandom rng = new SecureRandom();
        rng.nextBytes(aesKey);
        rng.nextBytes(nonce);
        Cipher gcm = Cipher.getInstance("AES/GCM/NoPadding");
        SecretKey key = new SecretKeySpec(aesKey, "AES");
        gcm.init(Cipher.ENCRYPT_MODE, key, new GCMParameterSpec(128, nonce));
        gcm.updateAAD(("1|" + kid).getBytes(StandardCharsets.UTF_8));
        byte[] sealed = gcm.doFinal(plainJson.getBytes(StandardCharsets.UTF_8));
        byte[] ciphertext = java.util.Arrays.copyOf(sealed, sealed.length - 16);
        byte[] tag = java.util.Arrays.copyOfRange(sealed, sealed.length - 16, sealed.length);
        Cipher rsa = Cipher.getInstance("RSA/ECB/OAEPPadding");
        OAEPParameterSpec oaep = new OAEPParameterSpec(
            "SHA-256", "MGF1", MGF1ParameterSpec.SHA256, PSource.PSpecified.DEFAULT);
        rsa.init(Cipher.ENCRYPT_MODE, pub, oaep);
        byte[] wrapped = new byte[44];
        System.arraycopy(aesKey, 0, wrapped, 0, 32);
        System.arraycopy(nonce, 0, wrapped, 32, 12);
        byte[] encKey = rsa.doFinal(wrapped);
        Base64.Encoder b64 = Base64.getEncoder();
        return "{\\"v\\":1,\\"kid\\":\\"" + jsonEscape(kid)
            + "\\",\\"enc_key\\":\\"" + b64.encodeToString(encKey)
            + "\\",\\"ciphertext\\":\\"" + b64.encodeToString(ciphertext)
            + "\\",\\"tag\\":\\"" + b64.encodeToString(tag) + "\\"}";
    }

    static String encryptCert(String password, byte[] p12, byte[] mp) throws Exception {
        // 步骤 2：提交加密结果，获取加密证书
        Base64.Encoder b64 = Base64.getEncoder();
        String plain = "{\\"password\\":\\"" + jsonEscape(password)
            + "\\",\\"p12_base64\\":\\"" + b64.encodeToString(p12)
            + "\\",\\"mp_base64\\":\\"" + b64.encodeToString(mp) + "\\"}";
        String body = encryptPayload(KID, PUB_PEM, plain);
        HttpRequest req = HttpRequest.newBuilder(URI.create(URL))
            .timeout(Duration.ofSeconds(90))
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        String raw = HttpClient.newHttpClient().send(req, HttpResponse.BodyHandlers.ofString()).body();
        int i = raw.indexOf("\\"mango\\"");
        int q1 = raw.indexOf('"', i + 7) + 1;
        int q2 = raw.indexOf('"', q1);
        return raw.substring(q1, q2);
    }

    public static void main(String[] args) throws Exception {
        // 步骤 1：准备明文数据
        byte[] p12 = Files.readAllBytes(Path.of("cert.p12"));
        byte[] mp = Files.readAllBytes(Path.of("cert.mobileprovision"));
        // 步骤 2：获取加密证书
        String mango = encryptCert("P12密码", p12, mp);
        // 步骤 3：把返回的 mango 做 Base64 解码，写入 APP 对应目录文件
        // 解压的 APP 目录（如 /xxxx/Payload/HillmountPlatform.app）
        Path appDir = Path.of("/xxxx/Payload/HillmountPlatform.app");
        Files.createDirectories(appDir.resolve("certificate"));
        Files.write(appDir.resolve("certificate/cert.mango"), Base64.getDecoder().decode(mango));
    }
}
`
