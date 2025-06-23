package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lizhening/WtfBackup/pkg/logger"
	"github.com/lizhening/WtfBackup/pkg/progress"
)

// FileOperator 文件操作接口
type FileOperator interface {
	Copy(src, dst string) error
	CopyWithProgress(src, dst string) error
	Walk(root string, walkFn filepath.WalkFunc) error
	EnsureDir(path string) error
	GetFileSize(path string) (int64, error)
	CopyDir(src, dst string, showProgress bool) error
	GetDirSize(path string) (int64, error)
	CleanOldBackups(backupDir string, keepCount int) error
}

// DefaultFileOperator 默认文件操作实现
type DefaultFileOperator struct {
	bufferSize int64
}

// NewDefaultFileOperator 创建默认文件操作器
func NewDefaultFileOperator(bufferSize int64) *DefaultFileOperator {
	if bufferSize <= 0 {
		bufferSize = 32 * 1024 // 默认32KB
	}
	return &DefaultFileOperator{
		bufferSize: bufferSize,
	}
}

// Copy 复制文件
func (op *DefaultFileOperator) Copy(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %w", err)
	}

	// 确保目标目录存在
	if err := op.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	buffer := make([]byte, op.bufferSize)
	_, err = io.CopyBuffer(dstFile, srcFile, buffer)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	return nil
}

// CopyWithProgress 带进度显示的复制文件
func (op *DefaultFileOperator) CopyWithProgress(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %w", err)
	}

	// 确保目标目录存在
	if err := op.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	// 创建进度写入器
	progressWriter := progress.NewProgressWriter(dstFile, srcInfo.Size(), "复制文件", filepath.Base(src))
	buffer := make([]byte, op.bufferSize)
	_, err = io.CopyBuffer(progressWriter, srcFile, buffer)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	fmt.Println() // 换行
	return nil
}

// Walk 遍历目录
func (op *DefaultFileOperator) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

// EnsureDir 确保目录存在
func (op *DefaultFileOperator) EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %w", path, err)
	}
	return nil
}

// GetFileSize 获取文件大小
func (op *DefaultFileOperator) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("获取文件信息失败: %w", err)
	}
	return info.Size(), nil
}

// CopyDir 复制目录
func (op *DefaultFileOperator) CopyDir(src, dst string, showProgress bool) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// 获取源目录信息
	_, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("获取源目录信息失败: %w", err)
	}

	// 创建目标目录
	if err := op.EnsureDir(dst); err != nil {
		return err
	}

	// 遍历源目录
	err = op.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %w", err)
		}

		// 构建目标路径
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// 创建目录
			if err := op.EnsureDir(dstPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			wg.Add(1)
			go func() {
				defer wg.Done()
				var err error
				if showProgress {
					err = op.CopyWithProgress(path, dstPath)
				} else {
					err = op.Copy(path, dstPath)
				}
				if err != nil {
					select {
					case errChan <- err:
					default:
					}
				}
			}()
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %w", err)
	}

	// 等待所有文件复制完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

// GetDirSize 获取目录大小
func (op *DefaultFileOperator) GetDirSize(path string) (int64, error) {
	var size int64
	err := op.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// CleanOldBackups 清理旧备份
func (op *DefaultFileOperator) CleanOldBackups(backupDir string, keepCount int) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("读取备份目录失败: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name()[:4] == "WTF_" {
			backups = append(backups, filepath.Join(backupDir, entry.Name()))
		}
	}

	if len(backups) <= keepCount {
		return nil
	}

	// 按修改时间排序
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			infoI, _ := os.Stat(backups[i])
			infoJ, _ := os.Stat(backups[j])
			if infoI.ModTime().Before(infoJ.ModTime()) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	// 删除旧备份
	for _, backup := range backups[keepCount:] {
		logger.Info("删除旧备份: %s", backup)
		if err := os.RemoveAll(backup); err != nil {
			logger.Error("删除备份失败 %s: %v", backup, err)
		}
	}

	return nil
}
