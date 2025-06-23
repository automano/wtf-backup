package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/lizhening/WtfBackup/backup"
	"github.com/lizhening/WtfBackup/config"
	"github.com/lizhening/WtfBackup/pkg/fileutil"
	"github.com/lizhening/WtfBackup/pkg/logger"
	"github.com/lizhening/WtfBackup/restore"
)

func main() {
	// 初始化日志系统
	log := logger.NewLogger(logger.LogLevelInfo, os.Stdout, "WTF-Backup")
	logger.SetDefaultLogger(log)

	// 初始化文件操作器
	fileOp := fileutil.NewDefaultFileOperator(32 * 1024) // 32KB buffer

	// 先加载配置文件
	configPath := config.DefaultConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Error("加载配置文件失败: %v", err)
		os.Exit(1)
	}

	// 创建子命令
	backupCmd := flag.NewFlagSet("backup", flag.ExitOnError)
	restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)

	// 备份命令参数 - 可选，如果不提供将使用配置文件中的设置
	wtfPath := backupCmd.String("wtf", cfg.WtfPath, "WTF文件夹路径 (可选，默认使用配置文件)")
	backupDir := backupCmd.String("backup", cfg.BackupDir, "备份保存的文件夹路径 (可选，默认使用配置文件)")
	showProgress := backupCmd.Bool("progress", true, "显示进度条")
	keepBackups := backupCmd.Int("keep", 5, "保留的备份数量")

	// 恢复命令参数
	restoreWtfPath := restoreCmd.String("wtf", cfg.WtfPath, "要恢复到的WTF文件夹路径 (可选，默认使用配置文件)")
	restoreBackupDir := restoreCmd.String("backup", cfg.BackupDir, "备份文件夹路径 (可选，默认使用配置文件)")
	addonName := restoreCmd.String("addon", "", "要恢复的插件名称 (可选，如不提供则恢复配置中的所有插件)")
	restoreShowProgress := restoreCmd.Bool("progress", true, "显示进度条")

	// 配置命令参数
	configWtfPath := configCmd.String("wtf", "", "设置WTF文件夹路径")
	configBackupDir := configCmd.String("backup", "", "设置备份文件夹路径")
	configAddAddons := configCmd.String("add-addons", "", "添加插件到恢复列表 (多个插件用逗号分隔)")
	configRemoveAddons := configCmd.String("remove-addons", "", "从恢复列表移除插件 (多个插件用逗号分隔)")
	configShowFlag := configCmd.Bool("show", false, "显示当前配置")

	// 检查参数
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 根据子命令执行不同的功能
	switch os.Args[1] {
	case "backup":
		backupCmd.Parse(os.Args[2:])
		// 更新配置
		if *wtfPath != "" && *wtfPath != cfg.WtfPath {
			cfg.WtfPath = config.NormalizePath(*wtfPath)
		}
		if *backupDir != "" && *backupDir != cfg.BackupDir {
			cfg.BackupDir = config.NormalizePath(*backupDir)
		}

		// 保存更新后的配置
		if err := config.SaveConfig(cfg, configPath); err != nil {
			logger.Error("保存配置文件失败: %v", err)
		}

		if cfg.WtfPath == "" || cfg.BackupDir == "" {
			logger.Error("必须提供WTF文件夹路径和备份路径，可以通过命令行参数或配置文件设置")
			backupCmd.PrintDefaults()
			os.Exit(1)
		}

		// 执行备份
		logger.Info("开始备份WTF文件夹...")
		err := backup.BackupWtf(*cfg, fileOp, *showProgress)
		if err != nil {
			logger.Error("备份失败: %v", err)
			os.Exit(1)
		}
		logger.Info("备份成功完成!")

		// 清理旧备份
		if *keepBackups > 0 {
			logger.Info("清理旧备份...")
			if err := fileOp.CleanOldBackups(cfg.BackupDir, *keepBackups); err != nil {
				logger.Error("清理旧备份失败: %v", err)
			}
		}

	case "restore":
		restoreCmd.Parse(os.Args[2:])
		// 更新配置
		if *restoreWtfPath != "" && *restoreWtfPath != cfg.WtfPath {
			cfg.WtfPath = config.NormalizePath(*restoreWtfPath)
		}
		if *restoreBackupDir != "" && *restoreBackupDir != cfg.BackupDir {
			cfg.BackupDir = config.NormalizePath(*restoreBackupDir)
		}

		// 保存更新后的配置
		if err := config.SaveConfig(cfg, configPath); err != nil {
			logger.Error("保存配置文件失败: %v", err)
		}

		if cfg.WtfPath == "" || cfg.BackupDir == "" {
			logger.Error("必须提供WTF文件夹路径和备份路径，可以通过命令行参数或配置文件设置")
			restoreCmd.PrintDefaults()
			os.Exit(1)
		}

		// 如果提供了插件名，则只恢复该插件
		if *addonName != "" {
			logger.Info("开始恢复插件 %s...", *addonName)
			err := restore.RestoreAddon(*cfg, *addonName, fileOp, *restoreShowProgress)
			if err != nil {
				logger.Error("恢复插件 %s 失败: %v", *addonName, err)
				os.Exit(1)
			}
			logger.Info("插件 %s 恢复成功完成!", *addonName)
		} else if len(cfg.Addons) > 0 {
			// 恢复配置中的所有插件
			logger.Info("将恢复配置中的 %d 个插件", len(cfg.Addons))
			for _, addon := range cfg.Addons {
				logger.Info("恢复插件: %s", addon)
				err := restore.RestoreAddon(*cfg, addon, fileOp, *restoreShowProgress)
				if err != nil {
					logger.Error("恢复插件 %s 失败: %v", addon, err)
					// 继续恢复其他插件
				} else {
					logger.Info("插件 %s 恢复成功!", addon)
				}
			}
			logger.Info("所有插件恢复操作完成!")
		} else {
			logger.Error("必须提供要恢复的插件名称，或在配置文件中配置插件列表")
			restoreCmd.PrintDefaults()
			os.Exit(1)
		}

	case "config":
		configCmd.Parse(os.Args[2:])

		// 更新WTF路径
		if *configWtfPath != "" {
			cfg.WtfPath = config.NormalizePath(*configWtfPath)
			logger.Info("已设置WTF路径: %s", cfg.WtfPath)
		}

		// 更新备份路径
		if *configBackupDir != "" {
			cfg.BackupDir = config.NormalizePath(*configBackupDir)
			logger.Info("已设置备份路径: %s", cfg.BackupDir)
		}

		// 添加插件
		if *configAddAddons != "" {
			addons := strings.Split(*configAddAddons, ",")
			for _, addon := range addons {
				addon = strings.TrimSpace(addon)
				if addon == "" {
					continue
				}

				// 检查是否已存在
				exists := false
				for _, a := range cfg.Addons {
					if a == addon {
						exists = true
						break
					}
				}

				if !exists {
					cfg.Addons = append(cfg.Addons, addon)
					logger.Info("已添加插件: %s", addon)
				} else {
					logger.Info("插件 %s 已在列表中", addon)
				}
			}
		}

		// 移除插件
		if *configRemoveAddons != "" {
			addons := strings.Split(*configRemoveAddons, ",")
			for _, addon := range addons {
				addon = strings.TrimSpace(addon)
				if addon == "" {
					continue
				}

				// 查找并移除
				for i, a := range cfg.Addons {
					if a == addon {
						cfg.Addons = append(cfg.Addons[:i], cfg.Addons[i+1:]...)
						logger.Info("已移除插件: %s", addon)
						break
					}
				}
			}
		}

		// 保存配置
		if err := config.SaveConfig(cfg, configPath); err != nil {
			logger.Error("保存配置文件失败: %v", err)
			os.Exit(1)
		}

		// 显示当前配置
		if *configShowFlag || (*configWtfPath == "" && *configBackupDir == "" && *configAddAddons == "" && *configRemoveAddons == "") {
			logger.Info("\n当前配置:")
			logger.Info("配置文件路径: %s", configPath)
			logger.Info("WTF文件夹路径: %s", cfg.WtfPath)
			logger.Info("备份文件夹路径: %s", cfg.BackupDir)
			logger.Info("插件列表:")
			if len(cfg.Addons) == 0 {
				logger.Info("  (无)")
			} else {
				for _, addon := range cfg.Addons {
					logger.Info("  - %s", addon)
				}
			}
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("WTF备份工具 - 备份和恢复魔兽世界的WTF文件夹")
	fmt.Println("\n用法:")
	fmt.Println("  backup: 备份WTF文件夹")
	fmt.Printf("    %s backup [-wtf <WTF文件夹路径>] [-backup <备份文件夹路径>] [-progress] [-keep <保留备份数量>]\n", os.Args[0])
	fmt.Println("  restore: 从备份中恢复插件配置")
	fmt.Printf("    %s restore [-wtf <WTF文件夹路径>] [-backup <备份文件夹路径>] [-addon <插件名称>] [-progress]\n", os.Args[0])
	fmt.Println("  config: 配置设置")
	fmt.Printf("    %s config [-wtf <WTF文件夹路径>] [-backup <备份文件夹路径>] [-add-addons <插件1,插件2...>] [-remove-addons <插件1,插件2...>] [-show]\n", os.Args[0])
}
