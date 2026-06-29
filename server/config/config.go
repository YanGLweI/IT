package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Upload   UploadConfig   `yaml:"upload"`
	LDAP     LDAPConfig     `yaml:"ldap"`
	Document DocumentConfig `yaml:"document"`
	RSA      RSAConfig      `yaml:"rsa"`
}

// DocumentConfig 文档配置
type DocumentConfig struct {
	LogoPath                        string `yaml:"logo_path"`
	PermissionDocumentVersion       string `yaml:"permission_document_version"`
	UserPermissionDocumentVersion  string `yaml:"user_permission_document_version"`
	AssetDocumentVersion           string `yaml:"asset_document_version"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port      string `yaml:"port"`
	JWTSecret string `yaml:"jwt_secret"`
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
	Path                          string `yaml:"path"`
	PolicyPath                    string `yaml:"policy_path"`
	TopologyPath                  string `yaml:"topology_path"`
	PermissionMonthlyCheckPath    string `yaml:"permission_monthly_check_path"`
	ThirdPartyQuarterlyCheckPath  string `yaml:"third_party_quarterly_check_path"`
	PermissionUserChangePath      string `yaml:"permission_user_change_path"`
	MaxSize                       int64  `yaml:"max_size"`
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

	Cfg = &Config{}
	if err := yaml.Unmarshal(data, Cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	if Cfg.Server.Port == "" {
		Cfg.Server.Port = ":8080"
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
	if Cfg.Upload.MaxSize == 0 {
		Cfg.Upload.MaxSize = 32 << 20
	}
	if Cfg.RSA.PrivateKeyPath == "" {
		Cfg.RSA.PrivateKeyPath = "./certificate/rsa_private.pem"
	}
	if Cfg.RSA.PublicKeyPath == "" {
		Cfg.RSA.PublicKeyPath = "./certificate/rsa_public.pem"
	}

	return nil
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
