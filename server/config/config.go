package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Database      DatabaseConfig      `yaml:"database"`
	Upload        UploadConfig        `yaml:"upload"`
	LDAP          LDAPConfig          `yaml:"ldap"`
	Document      DocumentConfig      `yaml:"document"`
	RSA           RSAConfig           `yaml:"rsa"`
	PasswordVault PasswordVaultConfig `yaml:"password_vault"`
}

// PasswordVaultConfig 密码本配置
type PasswordVaultConfig struct {
	AESKey string `yaml:"aes_key"`
}

// DocumentConfig 文档配置
type DocumentConfig struct {
	LogoPath                       string `yaml:"logo_path"`
	PermissionDocumentVersion      string `yaml:"permission_document_version"`
	UserPermissionDocumentVersion  string `yaml:"user_permission_document_version"`
	AssetDocumentVersion           string `yaml:"asset_document_version"`
	SystemHardeningDocumentVersion string `yaml:"system_hardening_document_version"`
}

// TLSConfig TLS/HTTPS 配置
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertPath string `yaml:"cert_path"`
	KeyPath  string `yaml:"key_path"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port               string    `yaml:"port"`
	JWTSecret          string    `yaml:"jwt_secret"`
	GinMode            string    `yaml:"gin_mode"`
	AccessTokenExpiry  int       `yaml:"access_token_expiry"`  // 分钟，默认120
	RefreshTokenExpiry int       `yaml:"refresh_token_expiry"` // 天，默认7
	TLS                TLSConfig `yaml:"tls"`
	ConfigKey          string    `yaml:"config-key"` // 配置加密私钥路径
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"dbname"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	Path                         string `yaml:"path"`
	PolicyPath                   string `yaml:"policy_path"`
	TopologyPath                 string `yaml:"topology_path"`
	PermissionMonthlyCheckPath   string `yaml:"permission_monthly_check_path"`
	ThirdPartyQuarterlyCheckPath string `yaml:"third_party_quarterly_check_path"`
	PermissionUserChangePath     string `yaml:"permission_user_change_path"`
	ChangeRecordTemplatePath     string `yaml:"change_record_template_path"`
	ChangeRecordPath             string `yaml:"change_record_path"`
	VulnerabilityScanPath        string `yaml:"vulnerability_scan_path"`
	SystemHardeningCheckPath     string `yaml:"system_hardening_check_path"`
	PenetrationTestPath          string `yaml:"penetration_test_path"`
	FirewallCheckPath            string `yaml:"firewall_check_path"`
	PatchUpdatePath              string `yaml:"patch_update_path"`
	BackupPath                   string `yaml:"backup_path"`
	BackupRecoveryPath           string `yaml:"backup_recovery_path"`
	BackupTemplatePath           string `yaml:"backup_template_path"`
	FormVaultPath                string `yaml:"form_vault_path"`
	FormVaultSnapshotPath        string `yaml:"form_vault_snapshot_path"`
	ITGuideMediaPath             string `yaml:"it_guide_media_path"`
}

// LDAPConfig LDAP配置
type LDAPConfig struct {
	Server          string `yaml:"server"`
	BaseDN          string `yaml:"base_dn"`
	UseTLS          bool   `yaml:"use_tls"`
	Insecure        bool   `yaml:"insecure"`
	UserFilter      string `yaml:"user_filter"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	SecurityGroupDN string `yaml:"security_group_dn"`
	CertPath        string `yaml:"cert_path"`
}

// RSAConfig RSA加密配置
type RSAConfig struct {
	PrivateKeyPath string `yaml:"private_key_path"`
	PublicKeyPath  string `yaml:"public_key_path"`
}

var Cfg *Config

// LoadConfig 加载配置文件
func LoadConfig() error {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 阶段一：提取 config-key 路径
	var raw struct {
		Server struct {
			ConfigKey string `yaml:"config-key"`
		} `yaml:"server"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 阶段二：解析为 yaml.Node 树，解密 ENC[...] 值
	var rootNode yaml.Node
	if err := yaml.Unmarshal(data, &rootNode); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	configKeyPath := raw.Server.ConfigKey
	if configKeyPath == "" {
		configKeyPath = "./key/config-key.pem"
	}

	// 加载私钥并解密
	privateKey, err := loadConfigPrivateKey(configKeyPath)
	if err != nil {
		return fmt.Errorf("加载配置解密密钥失败: %v", err)
	}
	if err := decryptEncValues(&rootNode, privateKey); err != nil {
		return fmt.Errorf("解密配置失败: %v", err)
	}

	// 将解密后的节点树解码到 Config 结构体
	Cfg = &Config{}
	if err := rootNode.Decode(Cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	if Cfg.Server.Port == "" {
		Cfg.Server.Port = ":8080"
	}
	if Cfg.Server.GinMode == "" {
		Cfg.Server.GinMode = "debug"
	}
	if Cfg.Server.AccessTokenExpiry == 0 {
		Cfg.Server.AccessTokenExpiry = 120
	}
	if Cfg.Server.RefreshTokenExpiry == 0 {
		Cfg.Server.RefreshTokenExpiry = 7
	}
	if Cfg.Database.DBName == "" {
		Cfg.Database.DBName = "it_platform"
	}
	if Cfg.Upload.Path == "" {
		Cfg.Upload.Path = "./uploads"
	}
	if Cfg.Upload.PolicyPath == "" {
		Cfg.Upload.PolicyPath = "./uploads/policies"
	}
	if Cfg.Upload.TopologyPath == "" {
		Cfg.Upload.TopologyPath = "./uploads/topologies"
	}
	if Cfg.Upload.PermissionMonthlyCheckPath == "" {
		Cfg.Upload.PermissionMonthlyCheckPath = "./uploads/permission_monthly_checks"
	}
	if Cfg.Upload.ThirdPartyQuarterlyCheckPath == "" {
		Cfg.Upload.ThirdPartyQuarterlyCheckPath = "./uploads/third_party_quarterly_checks"
	}
	if Cfg.Upload.PermissionUserChangePath == "" {
		Cfg.Upload.PermissionUserChangePath = "./uploads/permission_user_changes"
	}
	if Cfg.Upload.ChangeRecordTemplatePath == "" {
		Cfg.Upload.ChangeRecordTemplatePath = "./uploads/change_record_templates"
	}
	if Cfg.Upload.ChangeRecordPath == "" {
		Cfg.Upload.ChangeRecordPath = "./uploads/change_records"
	}
	if Cfg.Upload.VulnerabilityScanPath == "" {
		Cfg.Upload.VulnerabilityScanPath = "./uploads/vulnerability_scans"
	}
	if Cfg.Upload.SystemHardeningCheckPath == "" {
		Cfg.Upload.SystemHardeningCheckPath = "./uploads/system_hardening_checks"
	}
	if Cfg.Upload.PenetrationTestPath == "" {
		Cfg.Upload.PenetrationTestPath = "./uploads/penetration_tests"
	}
	if Cfg.Upload.FirewallCheckPath == "" {
		Cfg.Upload.FirewallCheckPath = "./uploads/firewall_checks"
	}
	if Cfg.Upload.PatchUpdatePath == "" {
		Cfg.Upload.PatchUpdatePath = "./uploads/patch_updates"
	}
	if Cfg.Upload.BackupPath == "" {
		Cfg.Upload.BackupPath = "./uploads/backups"
	}
	if Cfg.Upload.BackupRecoveryPath == "" {
		Cfg.Upload.BackupRecoveryPath = "./uploads/backup_recoveries"
	}
	if Cfg.Upload.BackupTemplatePath == "" {
		Cfg.Upload.BackupTemplatePath = "./uploads/backup_templates"
	}
	if Cfg.Upload.FormVaultPath == "" {
		Cfg.Upload.FormVaultPath = "./uploads/form_vault"
	}
	if Cfg.Upload.FormVaultSnapshotPath == "" {
		Cfg.Upload.FormVaultSnapshotPath = "./uploads/form_vault/snapshots"
	}
	if Cfg.Upload.ITGuideMediaPath == "" {
		Cfg.Upload.ITGuideMediaPath = "./uploads/it_guide_media"
	}
	if Cfg.RSA.PrivateKeyPath == "" {
		Cfg.RSA.PrivateKeyPath = "./certificate/rsa_private.pem"
	}
	if Cfg.RSA.PublicKeyPath == "" {
		Cfg.RSA.PublicKeyPath = "./certificate/rsa_public.pem"
	}

	if Cfg.Server.TLS.CertPath == "" {
		Cfg.Server.TLS.CertPath = "./certificate/server.crt"
	}
	if Cfg.Server.TLS.KeyPath == "" {
		Cfg.Server.TLS.KeyPath = "./certificate/server.key"
	}

	return nil
}

// loadConfigPrivateKey 读取配置解密用的 RSA 私钥
func loadConfigPrivateKey(path string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %v", err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("解析私钥PEM失败")
	}
	// 支持 PKCS#1 和 PKCS#8 两种格式
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	keyIface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析RSA私钥失败")
	}
	rsaKey, ok := keyIface.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("私钥不是RSA类型")
	}
	return rsaKey, nil
}

// decryptEncValues 递归遍历 yaml.Node 树，解密所有 ENC[base64] 格式的值
func decryptEncValues(node *yaml.Node, key *rsa.PrivateKey) error {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.ScalarNode {
		val := node.Value
		if strings.HasPrefix(val, "ENC[") && strings.HasSuffix(val, "]") {
			cipherBase64 := val[4 : len(val)-1]
			plainText, err := rsaDecryptOAEP(cipherBase64, key)
			if err != nil {
				return fmt.Errorf("解密字段失败: %v", err)
			}
			node.Value = plainText
		}
		return nil
	}
	for _, child := range node.Content {
		if err := decryptEncValues(child, key); err != nil {
			return err
		}
	}
	return nil
}

// rsaDecryptOAEP 使用 RSA-OAEP + SHA-256 解密 base64 编码的密文
func rsaDecryptOAEP(cipherBase64 string, key *rsa.PrivateKey) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}
	plainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, key, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("RSA解密失败: %v", err)
	}
	return string(plainText), nil
}

// GetDSN 获取数据库连接字符串
func GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		Cfg.Database.Username,
		Cfg.Database.Password,
		Cfg.Database.Host,
		Cfg.Database.Port,
		Cfg.Database.DBName,
	)
}
