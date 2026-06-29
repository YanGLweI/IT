package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

	"it-platform-server/config"

	"github.com/gin-gonic/gin"
)

var (
	rsaPrivateKey   *rsa.PrivateKey
	rsaPublicKeyPEM string
)

// InitRSAKeys 初始化RSA密钥对（在main.go中调用）
func InitRSAKeys() error {
	cfg := &config.Cfg.RSA

	// 读取私钥
	privPEM, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("读取RSA私钥失败: %v", err)
	}
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return fmt.Errorf("解析RSA私钥PEM失败")
	}
	// 支持PKCS#8和PKCS#1两种格式
	privKeyIface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privKeyIface, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("解析RSA私钥失败: %v", err)
		}
	}
	rsaPrivateKey = privKeyIface.(*rsa.PrivateKey)

	// 读取公钥
	pubPEM, err := os.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("读取RSA公钥失败: %v", err)
	}
	block, _ = pem.Decode(pubPEM)
	if block == nil {
		return fmt.Errorf("解析RSA公钥PEM失败")
	}
	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析RSA公钥失败: %v", err)
	}

	// 缓存公钥PEM字符串
	rsaPublicKeyPEM = string(pubPEM)

	fmt.Println("RSA密钥对加载成功")
	return nil
}

// GetPublicKey 获取RSA公钥（公开接口，无需认证）
func GetPublicKey(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"public_key": rsaPublicKeyPEM,
		},
	})
}

// DecryptPassword 使用RSA私钥解密base64编码的密文
func DecryptPassword(encryptedBase64 string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	// 使用OAEP + SHA-256解密
	plainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPrivateKey, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("RSA解密失败: %v", err)
	}

	return string(plainText), nil
}
