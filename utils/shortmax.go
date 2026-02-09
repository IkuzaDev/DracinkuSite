package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
)

var shortmaxPublicKey = []byte{
	77, 73, 71, 102, 77, 65, 48, 71, 67, 83, 113, 71, 83, 73, 98, 51, 68, 81,
	69, 66, 65, 81, 85, 65, 65, 52, 71, 78, 65, 68, 67, 66, 105, 81, 75, 66,
	103, 81, 68, 81, 78, 56, 119, 74, 69, 118, 100, 99, 112, 107, 108, 107,
	117, 51, 117, 119, 120, 76, 97, 122, 54, 55, 79, 116, 80, 51, 71, 112,
	65, 75, 50, 65, 72, 85, 89, 47, 43, 80, 87, 118, 105, 106, 80, 86, 75,
	66, 83, 82, 67, 88, 68, 54, 119, 105, 114, 43, 74, 75, 74, 81, 55, 122,
	99, 120, 55, 115, 69, 110, 65, 120, 85, 105, 47, 67, 77, 122, 119, 82,
	110, 43, 66, 109, 75, 111, 109, 73, 118, 81, 112, 122, 73, 85, 122, 76,
	117, 97, 65, 120, 105, 117, 53, 82, 77, 48, 66, 65, 107, 75, 87, 88, 65,
	120, 75, 83, 99, 115, 106, 56, 110, 89, 51, 118, 108, 98, 67, 104, 79,
	100, 119, 52, 73, 67, 120, 113, 114, 83, 79, 51, 81, 90, 77, 101, 68,
	120, 122, 121, 78, 76, 88, 75, 110, 88, 109, 55, 79, 70, 66, 66, 47,
	119, 53, 120, 108, 105, 53, 122, 85, 104, 68, 81, 73, 68, 65, 81, 65, 66,
}

func RC4MD5(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

func RC4Crypt(data []byte, key string) []byte {
	s := make([]byte, 256)
	for i := 0; i < 256; i++ {
		s[i] = byte(i)
	}

	j := 0
	for i := 0; i < 256; i++ {
		j = (j + int(s[i]) + int(key[i%len(key)])) & 0xff
		s[i], s[j] = s[j], s[i]
	}

	out := make([]byte, len(data))
	ii := 0
	j = 0

	for k := 0; k < len(data); k++ {
		ii = (ii + 1) & 0xff
		j = (j + int(s[ii])) & 0xff
		s[ii], s[j] = s[j], s[ii]
		out[k] = data[k] ^ s[(int(s[ii])+int(s[j]))&0xff]
	}

	return out
}

func RC4Encrypt(plainText, key string) string {
	if plainText == "" || key == "" {
		return ""
	}
	encrypted := RC4Crypt([]byte(plainText), key)
	return hex.EncodeToString(encrypted)
}

func RC4Decrypt(hexText, key string) string {
	if hexText == "" || key == "" {
		return ""
	}
	data, err := hex.DecodeString(hexText)
	if err != nil {
		return ""
	}
	decrypted := RC4Crypt(data, key)
	return string(decrypted)
}

func RC4RSAEncrypt(data string) (string, error) {
	pubKeyBase64 := string(shortmaxPublicKey)
	derBytes, err := base64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return "", err
	}

	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return "", err
	}

	publicKey := pub.(*rsa.PublicKey)
	chunkSize := 117
	var encryptedChunks []byte

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := []byte(data[i:end])
		encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, chunk)
		if err != nil {
			return "", err
		}
		encryptedChunks = append(encryptedChunks, encrypted...)
	}

	return base64.StdEncoding.EncodeToString(encryptedChunks), nil
}
