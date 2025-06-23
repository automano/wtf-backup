package restore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lizhening/WtfBackup/config"
	"github.com/lizhening/WtfBackup/pkg/fileutil"
	"github.com/lizhening/WtfBackup/pkg/logger"
)

// RestoreAddon 从备份中恢复特定插件的配置
func RestoreAddon(cfg config.Config, addonName string, fileOp fileutil.FileOperator, showProgress bool) error {
	// 找到最新的备份
	backups, err := findBackups(cfg.BackupDir)
	if err != nil {
		return fmt.Errorf("查找备份失败: %w", err)
	}
	if len(backups) == 0 {
		return fmt.Errorf("没有找到备份")
	}

	// 获取最新的备份
	latestBackup := backups[0]
	logger.Info("将从备份 %s 中恢复插件 %s 的配置", filepath.Base(latestBackup), addonName)

	// 准备查找插件相关的文件夹和文件
	// WTF文件夹通常有以下与插件相关的路径：
	// 1. Account/<账号>/SavedVariables/<插件名>.lua
	// 2. Account/<账号>/<服务器>/<角色>/SavedVariables/<插件名>.lua
	// 3. Account/<账号>/SavedVariablesPerCharacter/<插件名>.lua
	// 4. Account/<账号>/<服务器>/<角色>/SavedVariablesPerCharacter/<插件名>.lua

	// 遍历备份文件夹找到所有与插件相关的配置
	err = fileOp.Walk(latestBackup, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(latestBackup, path)
		if err != nil {
			return err
		}

		// 检查是否是插件相关文件
		if isAddonFile(relPath, addonName) {
			// 构建目标路径
			destPath := filepath.Join(cfg.WtfPath, relPath)
			destDir := filepath.Dir(destPath)

			// 创建必要的文件夹
			if err := fileOp.EnsureDir(destDir); err != nil {
				return fmt.Errorf("创建文件夹 %s 失败: %w", destDir, err)
			}

			// 如果是文件，则复制
			if !info.IsDir() {
				var err error
				if showProgress {
					err = fileOp.CopyWithProgress(path, destPath)
				} else {
					err = fileOp.Copy(path, destPath)
				}
				if err != nil {
					return fmt.Errorf("复制文件 %s 至 %s 失败: %w", path, destPath, err)
				}
				logger.Info("已恢复: %s", relPath)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("恢复过程中出错: %w", err)
	}

	return nil
}

// isAddonFile 检查文件是否与指定插件相关
func isAddonFile(relPath, addonName string) bool {
	// 这些是插件配置文件的常见位置
	patterns := []string{
		// 全局设置
		fmt.Sprintf("Account/*/SavedVariables/%s.lua", addonName),
		// 角色特定设置
		fmt.Sprintf("Account/*/*/*/SavedVariables/%s.lua", addonName),
		// 角色特定设置 (另一种类型)
		fmt.Sprintf("Account/*/SavedVariablesPerCharacter/%s.lua", addonName),
		// 角色特定设置 (另一种类型)
		fmt.Sprintf("Account/*/*/*/SavedVariablesPerCharacter/%s.lua", addonName),
	}

	// 将路径分隔符统一为 '/'
	relPath = filepath.ToSlash(relPath)

	// 检查文件是否匹配任何模式
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
	}

	// 处理可能的子文件夹和其他相关文件
	if strings.Contains(relPath, fmt.Sprintf("/SavedVariables/%s_", addonName)) ||
		strings.Contains(relPath, fmt.Sprintf("/SavedVariablesPerCharacter/%s_", addonName)) {
		return true
	}

	return false
}

// findBackups 查找并按时间排序所有备份
func findBackups(backupDir string) ([]string, error) {
	// 确保备份目录存在
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("备份目录不存在")
	}

	// 获取备份目录下的所有条目
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	// 过滤出备份文件夹
	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "WTF_Backup_") {
			backups = append(backups, filepath.Join(backupDir, entry.Name()))
		}
	}

	// 按时间倒序排列（最新的备份在最前面）
	// 由于备份文件夹的命名格式中包含时间戳，所以可以直接按名称排序
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if filepath.Base(backups[i]) < filepath.Base(backups[j]) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	return backups, nil
}
