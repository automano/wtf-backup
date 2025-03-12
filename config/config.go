package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 保存程序配置
type Config struct {
	// WTF文件夹路径
	WtfPath string `yaml:"wtf_path"`
	// 备份文件夹路径
	BackupDir string `yaml:"backup_dir"`
	// 需要恢复的插件列表
	Addons []string `yaml:"addons"`
}

// DefaultConfigPath 返回默认配置文件路径
func DefaultConfigPath() string {
	// 使用当前工作目录作为项目目录
	workDir, err := os.Getwd()
	if err != nil {
		return "config.yaml" // 如果无法获取当前目录，使用相对路径
	}
	return filepath.Join(workDir, "config.yaml")
}

// LoadConfig 从配置文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			WtfPath:   "",
			BackupDir: "",
			Addons:    []string{},
		}, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 规范化路径
	config.WtfPath = NormalizePath(config.WtfPath)
	config.BackupDir = NormalizePath(config.BackupDir)

	return &config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	// 将路径规范化后再保存
	configToSave := &Config{
		WtfPath:   NormalizePath(config.WtfPath),
		BackupDir: NormalizePath(config.BackupDir),
		Addons:    config.Addons,
	}

	// 将配置序列化为YAML
	data, err := yaml.Marshal(configToSave)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 创建配置文件目录
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// NormalizePath 规范化路径，确保路径格式符合当前操作系统
func NormalizePath(path string) string {
	if path == "" {
		return path
	}

	// 将波浪号(~)替换为用户主目录
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	}

	// 将正斜杠(/)转换为系统分隔符
	path = filepath.FromSlash(path)

	// 清理路径（去除多余的分隔符等）
	path = filepath.Clean(path)

	// 在Windows上，确保路径使用反斜杠(\)
	if runtime.GOOS == "windows" {
		// 如果是相对路径且没有驱动器前缀，不修改
		if !filepath.IsAbs(path) && !strings.Contains(path, ":") {
			return path
		}

		// 确保Windows路径使用反斜杠
		path = strings.ReplaceAll(path, "/", "\\")

		// 确保驱动器盘符大写
		if len(path) >= 2 && path[1] == ':' {
			path = strings.ToUpper(path[:1]) + path[1:]
		}
	}

	return path
}
