package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lizhening/WtfBackup/config"
	"github.com/lizhening/WtfBackup/pkg/fileutil"
	"github.com/lizhening/WtfBackup/pkg/logger"
)

// BackupWtf 备份WTF文件夹
func BackupWtf(cfg config.Config, fileOp fileutil.FileOperator, showProgress bool) error {
	// 验证WTF文件夹存在
	info, err := os.Stat(cfg.WtfPath)
	if err != nil {
		return fmt.Errorf("无法访问WTF文件夹: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是一个文件夹", cfg.WtfPath)
	}

	// 生成备份文件夹名称，格式为 WTF_Backup_YYYY-MM-DD_HH-MM-SS
	now := time.Now()
	backupName := fmt.Sprintf("WTF_Backup_%s", now.Format("2006-01-02_15-04-05"))
	backupPath := filepath.Join(cfg.BackupDir, backupName)

	// 创建备份文件夹
	if err := fileOp.EnsureDir(backupPath); err != nil {
		return fmt.Errorf("创建备份文件夹失败: %w", err)
	}

	// 开始复制文件
	logger.Info("开始备份WTF文件夹到: %s", backupPath)
	err = fileOp.CopyDir(cfg.WtfPath, backupPath, showProgress)
	if err != nil {
		return fmt.Errorf("备份过程中出错: %w", err)
	}

	return nil
}

// copyDir 递归复制文件夹内容
func copyDir(src, dst string) error {
	// 获取源文件夹信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标文件夹
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// 读取源文件夹中的内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 遍历并复制每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 递归复制子文件夹
			if err = copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err = copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 获取源文件的权限
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// 创建目标文件
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
