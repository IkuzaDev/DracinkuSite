package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"strings"
)

type NetShortUtil struct {
	publicKeyPem  string
	privateKeyPem string
}

func NewNetShortUtil() *NetShortUtil {
	return &NetShortUtil{
		publicKeyPem:  "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2poXMstZ8NCWE7915MXzDWC5/t+oB2waGfskPqSZwLqxd4ZBR0H1cb1tAZRZcV7P+LmOd6SYNxhnELaWuKTD+D3xkz8Tt1L5j/ynGqVt1MDbiQIEzXQKUkNDSH6T0A+Xzo/67/8QOQXlVJfW06resbaeNvibfx6Qc78j96bCIPlxPrtieilVTBHUFOXjirxK/ki/mO8P2smRbpt73fsQWdGmTGMfYGvfPApGyxbxLkL/qrBjU25XpM8a0MBqzFWUAchHmqSBJ6Mbfam1SSgf3b2U28s67nOW+JiOrhd6iVLcsLFxXA54HX+Zbej3AbOB6jKaEmp/bz1amneE1NYXwwIDAQAB",
		privateKeyPem: "MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCK0Tl1pd7bjTRU93bWoHW1hLCDj2+9bg1MgY8j5C7xXaw6bJfToXhWbH1fXNbnFFVqxyYNErcuOUwJZxyDgcxUXM4yWnRseb2GF97GOicAQ2keDzVYmwky4lrSRwvcXutJRLPUCRQNfc6upfk2G5TKh6/CcP4TV1eXTF7+vdEw2SHxAOITKbSfcaZXr/hVs6a1aRHsBF+7RG99ebwZIP6/AgIyqX9RbDVN6ixi1v2G3/bwAULHLSqGdSaqij/ca17fbFGITaeCeEaZ6d/P4ZuOK+PEPdbPQt6SbY4lZaYwRvdrpH73kigPITgDzIDONFybJ1m7wRKlq1wxWHwbimptAgMBAAECggEAPz3cYJXFtt5YphDrahJGLgEabYVOUc2ub1li/eX54OpdCWzpqneYnD7myyg/m5zu4SuDUVdibsOZuXrpSZw7m3+ATP5apgS8bDe5vTNHC16qqBAjrI9NHIp09/F4HNh9dq6/Am10XkUfgP+KTrU4DyDL2NijV+pltD8N1B5kDE1igokVcsavhnu2INoMRXYE78Wq6urNECuFWw9hldv81M9m2w56t1CQOUukpo4mfmLjZRe2s+kwtcBVefGHP8Cj0OeH2dGltjl2YSQMRBFUCVoixYpOrcjIHoqzWri8IfUZ2tW+nUvHl5IZ9RVxefnFaLGnxiXd2sk6Sn4aD/l9YQKBgQDVv3HaOZxHRqlNSPrNGqplGhE066HnDsq6MlPukiovxE43CRBmpTnk9zDCqrDh9t2HbJuao7nSq5WlBERWgwqXU/qDpH43W7Y/lJfHkDv6A2m0viJa0a9x8+CJpNnCDu1ATo4/IQKwoXYice6JKnUyXgkGKn+HipiN6tO0EtWHlQKBgQCmQfklKFtXtm/FZ6NIMs+d+EyvaE5xNLKGYQxmiCR10WGYd8ZV+K0Q6qXHS+a32TirWB9F3TqPOklTytMrfPZB3BCXj4weEldb8W716G8FYf7LLhaT+MdpF7KDcruObwoQAvKV3N4eX6tUEMmdrx9hpCmmIU5EeXUkhGdmwk7BeQKBgAIXMkThJV8pGMTRvuo8pYgBnkN3PoklAuSZU2rU8Sawc9dj9k4atZtAs7BjvQEoyffmHwt/KHUgCoGnrgdulq7uOlgJRtbBxeGPUYC5L2z9lY4YAfwDawThTsPp4dtdDAMCAbAqYX1axu4FUUD0MltAwjPWPJMVzvIsZs+vE3mVAoGAJPja3OaCmZjadj2709xoyypic0dw2j/ry3JdfZec9A5h87P/CTNJ2U81GoLIhe3qakAohDLUSPGfSOD74NnjMXYswmeLs0xE3Q9tq4XK2pmWPby8DJ/wSHCapByplN0gkbr2E1mQk5SW1xT8oPJGukH1eRpC+3s/D6XaEMH5HZECgYEAigoX5l39LDsCgeaUcI4S9grkaas/WsKv37eqo3oD9Qk6VFiMM5L5Zig6aXJxuAPLVjb38caJRPmPmOXLT2kEP1E1h6OJOhEhETwVIUtcBzsK25ju9LqL89bC+W0uS7BPvk6Tcws/tXHCkQCTgb9jVXceZ2ox+6axvlW/5WgHt5Q=",
	}
}

func (n *NetShortUtil) formatPem(key, keyType string) string {
	if strings.HasPrefix(key, "-----BEGIN") {
		return key
	}
	return "-----BEGIN " + keyType + "-----\n" + key + "\n-----END " + keyType + "-----"
}

func (n *NetShortUtil) RSAEncrypt(plain string) (string, error) {
	pemData := n.formatPem(n.publicKeyPem, "PUBLIC KEY")
	block, _ := decodePEM(pemData)
	if block == nil {
		return "", nil
	}

	pub, err := x509.ParsePKIXPublicKey(block)
	if err != nil {
		return "", err
	}

	publicKey := pub.(*rsa.PublicKey)
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(plain))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (n *NetShortUtil) RSADecrypt(ciphertext string) (string, error) {
	pemData := n.formatPem(n.privateKeyPem, "PRIVATE KEY")
	block, _ := decodePEM(pemData)
	if block == nil {
		return "", nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block)
	if err != nil {
		return "", err
	}

	privateKey := key.(*rsa.PrivateKey)
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, decoded)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func decodePEM(pemData string) ([]byte, error) {
	lines := strings.Split(pemData, "\n")
	var b64 strings.Builder
	for _, line := range lines {
		if strings.HasPrefix(line, "-----") {
			continue
		}
		b64.WriteString(strings.TrimSpace(line))
	}
	return base64.StdEncoding.DecodeString(b64.String())
}

func (n *NetShortUtil) normalizeKey(key string) []byte {
	keyBytes := []byte(key)
	if len(keyBytes) == 32 {
		return keyBytes
	}
	hash := sha256.Sum256(keyBytes)
	return hash[:]
}

func pkcs5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func pkcs5Unpadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}
	unpadding := int(data[length-1])
	if unpadding > length {
		return data
	}
	return data[:(length - unpadding)]
}

func (n *NetShortUtil) AESEncrypt(plain, key string) (string, error) {
	normalizedKey := n.normalizeKey(key)
	block, err := aes.NewCipher(normalizedKey)
	if err != nil {
		return "", err
	}

	plainBytes := pkcs5Padding([]byte(plain), aes.BlockSize)
	encrypted := make([]byte, len(plainBytes))

	for i := 0; i < len(plainBytes); i += aes.BlockSize {
		block.Encrypt(encrypted[i:i+aes.BlockSize], plainBytes[i:i+aes.BlockSize])
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (n *NetShortUtil) AESDecrypt(ciphertext, key string) (string, error) {
	normalizedKey := n.normalizeKey(key)
	block, err := aes.NewCipher(normalizedKey)
	if err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(decoded))
	for i := 0; i < len(decoded); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], decoded[i:i+aes.BlockSize])
	}

	return string(pkcs5Unpadding(decrypted)), nil
}

func (n *NetShortUtil) ToBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (n *NetShortUtil) FromBase64(b64 string) string {
	decoded, _ := base64.StdEncoding.DecodeString(b64)
	return string(decoded)
}

func (n *NetShortUtil) GenerateDeviceCode() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 30)
	rand.Read(b)

	result := make([]byte, 30)
	for i := 0; i < 30; i++ {
		result[i] = chars[int(b[i])%len(chars)]
	}

	return "G-" + strings.ToLower(string(result))
}
