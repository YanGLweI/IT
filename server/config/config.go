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
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
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
	Path         string `yaml:"path"`
	PolicyPath   string `yaml:"policy_path"`
	TopologyPath string `yaml:"topology_path"`
	MaxSize      int64  `yaml:"max_size"`
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
	if Cfg.Upload.MaxSize == 0 {
		Cfg.Upload.MaxSize = 32 << 20
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
