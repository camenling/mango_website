export const unblacklistPhpSample = `<?php
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
const URL = 'https://<本站域名>/api/sign/channel-unblacklist';

function encrypt_payload($kid, $pubPem, array $plainObj)
{
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

function remove_blacklist($udid)
{
    // 步骤 1：准备明文数据
    $plain = ['udid' => $udid];
    // 步骤 2：加密请求 POST https://<本站域名>/api/sign/channel-unblacklist
    $body = json_encode(encrypt_payload(KID, PUB_PEM, $plain));
    $ch = curl_init(URL);
    curl_setopt_array($ch, [
        CURLOPT_POST => true,
        CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
        CURLOPT_POSTFIELDS => $body,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_TIMEOUT => 90,
    ]);
    $raw = curl_exec($ch);
    $status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    // 步骤 3：解析返回
    return [$status, json_decode($raw, true)];
}

list($status, $json) = remove_blacklist('00008020-001234567890ABCD');
echo $status, PHP_EOL;
echo json_encode($json, JSON_UNESCAPED_UNICODE), PHP_EOL;
`

export const unblacklistGoSample = `package main

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
	apiURL = "https://<本站域名>/api/sign/channel-unblacklist"
)

func encryptPayload(kid, pubPEM string, plainObj map[string]string) (map[string]any, error) {
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

func removeBlacklist(udid string) (int, []byte, error) {
	// 步骤 1：准备明文数据
	plain := map[string]string{"udid": udid}
	env, err := encryptPayload(kid, pubPEM, plain)
	if err != nil {
		return 0, nil, err
	}
	body, err := json.Marshal(env)
	if err != nil {
		return 0, nil, err
	}
	// 步骤 2：加密请求 POST https://<本站域名>/api/sign/channel-unblacklist
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	// 步骤 3：解析返回
	return resp.StatusCode, raw, nil
}

func main() {
	status, raw, err := removeBlacklist("00008020-001234567890ABCD")
	if err != nil {
		panic(err)
	}
	fmt.Println(status)
	os.Stdout.Write(raw)
	fmt.Println()
}
`

export const unblacklistJavaSample = `// Java 11+
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
import java.security.KeyFactory;
import java.security.PublicKey;
import java.security.SecureRandom;
import java.security.spec.MGF1ParameterSpec;
import java.security.spec.X509EncodedKeySpec;
import java.time.Duration;
import java.util.Base64;

public class ChannelUnblacklist {
    static final String KID = "你的 kid";
    static final String PUB_PEM = "-----BEGIN PUBLIC KEY-----\\n你的公钥\\n-----END PUBLIC KEY-----\\n";
    static final String URL = "https://<本站域名>/api/sign/channel-unblacklist";

    static String jsonEscape(String s) {
        return s.replace("\\\\", "\\\\\\\\").replace("\\"", "\\\\\\"");
    }

    static String encryptPayload(String kid, String pubPem, String plainJson) throws Exception {
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

    static HttpResponse<String> removeBlacklist(String udid) throws Exception {
        // 步骤 1：准备明文数据
        String plain = "{\\"udid\\":\\"" + jsonEscape(udid) + "\\"}";
        // 步骤 2：加密请求 POST https://<本站域名>/api/sign/channel-unblacklist
        String body = encryptPayload(KID, PUB_PEM, plain);
        HttpRequest req = HttpRequest.newBuilder(URI.create(URL))
            .timeout(Duration.ofSeconds(90))
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        // 步骤 3：解析返回
        return HttpClient.newHttpClient().send(req, HttpResponse.BodyHandlers.ofString());
    }

    public static void main(String[] args) throws Exception {
        HttpResponse<String> resp = removeBlacklist("00008020-001234567890ABCD");
        System.out.println(resp.statusCode());
        System.out.println(resp.body());
    }
}
`
