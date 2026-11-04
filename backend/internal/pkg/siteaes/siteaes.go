package siteaes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Encrypt 与授权 libUtils.EncryptAES256 一致：MD5(key) + AES-CBC + IV||密文 Base64。
func Encrypt(key, plaintext string) (string, error) {
	aesKey := md5.Sum([]byte(key))
	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return "", err
	}
	plain := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	out := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, plain)
	return base64.StdEncoding.EncodeToString(append(iv, out...)), nil
}

// Decrypt 与授权 libUtils.DecryptAES256 一致。
func Decrypt(key, encryptedText string) (string, error) {
	aesKey := md5.Sum([]byte(key))
	raw, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}
	if len(raw) < aes.BlockSize {
		return "", fmt.Errorf("密文太短")
	}
	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return "", err
	}
	iv, ct := raw[:aes.BlockSize], raw[aes.BlockSize:]
	plain := make([]byte, len(ct))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ct)
	return string(pkcs7Unpad(plain)), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := make([]byte, padding)
	for i := range pad {
		pad[i] = byte(padding)
	}
	return append(data, pad...)
}

func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	n := int(data[len(data)-1])
	if n <= 0 || n > len(data) {
		return data
	}
	return data[:len(data)-n]
}
