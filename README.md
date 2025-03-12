# WTF Backup

这是一个用于备份和恢复魔兽世界 WTF 文件夹的工具。

## 功能

1. 完整备份 WTF 文件夹，无需压缩
2. 从备份中恢复特定插件的配置
3. 使用配置文件保存路径和插件列表，无需每次都输入

## 安装

```bash
git clone https://github.com/lizhening/WtfBackup.git
cd WtfBackup
go build
```

## 配置文件

程序使用 YAML 格式的配置文件保存以下设置：
- WTF 文件夹路径
- 备份文件夹路径
- 要恢复的插件列表

配置文件位于程序运行目录下的 `config.yaml`。

配置文件示例 (Linux/macOS):
```yaml
wtf_path: /path/to/World of Warcraft/_retail_/WTF
backup_dir: /path/to/backup/folder
addons:
  - DBM-Core
  - ElvUI
  - WeakAuras
```

配置文件示例 (Windows):
```yaml
wtf_path: C:\Games\World of Warcraft\_retail_\WTF
backup_dir: D:\WoW_Backups
addons:
  - DBM-Core
  - ElvUI
  - WeakAuras
```

### 设置配置

在 Linux/macOS 上:
```bash
# 设置 WTF 路径和备份路径
./WtfBackup config -wtf "/path/to/World of Warcraft/_retail_/WTF" -backup "/path/to/backup/folder"
```

在 Windows 上:
```bash
# 设置 WTF 路径和备份路径
WtfBackup.exe config -wtf "C:\Games\World of Warcraft\_retail_\WTF" -backup "D:\WoW_Backups"
```

```bash
# 添加插件到恢复列表
./WtfBackup config -add-addons "DBM-Core,Details,ElvUI"

# 从列表移除插件
./WtfBackup config -remove-addons "Details"

# 显示当前配置
./WtfBackup config -show
# 或简单地
./WtfBackup config
```

## 使用说明

### 备份 WTF 文件夹

Linux/macOS:
```bash
# 使用配置文件中的路径
./WtfBackup backup

# 或指定路径（会更新配置文件）
./WtfBackup backup -wtf "/path/to/World of Warcraft/_retail_/WTF" -backup "/path/to/backup/folder"
```

Windows:
```bash
# 使用配置文件中的路径
WtfBackup.exe backup

# 或指定路径（会更新配置文件）
WtfBackup.exe backup -wtf "C:\Games\World of Warcraft\_retail_\WTF" -backup "D:\WoW_Backups"
```

备份将存储在指定的备份文件夹中，以时间戳命名（例如 `WTF_Backup_2023-04-23_15-30-45`）。

### 恢复插件配置

Linux/macOS:
```bash
# 恢复指定的插件
./WtfBackup restore -addon "DBM-Core"

# 或恢复配置文件中的所有插件
./WtfBackup restore

# 也可以指定路径（会更新配置文件）
./WtfBackup restore -wtf "/path/to/World of Warcraft/_retail_/WTF" -backup "/path/to/backup/folder" -addon "DBM-Core"
```

Windows:
```bash
# 恢复指定的插件
WtfBackup.exe restore -addon "DBM-Core"

# 或恢复配置文件中的所有插件
WtfBackup.exe restore

# 也可以指定路径（会更新配置文件）
WtfBackup.exe restore -wtf "C:\Games\World of Warcraft\_retail_\WTF" -backup "D:\WoW_Backups" -addon "DBM-Core"
```

程序将从最新的备份中恢复指定插件或所有配置中的插件。

## 魔兽世界 WTF 文件夹结构

WTF 文件夹包含以下与插件相关的配置：

1. `Account/<账号>/SavedVariables/<插件名>.lua` - 账号级别的插件设置
2. `Account/<账号>/<服务器>/<角色>/SavedVariables/<插件名>.lua` - 角色级别的插件设置
3. `Account/<账号>/SavedVariablesPerCharacter/<插件名>.lua` - 账号级别的角色特定插件设置
4. `Account/<账号>/<服务器>/<角色>/SavedVariablesPerCharacter/<插件名>.lua` - 角色级别的角色特定插件设置

本程序会自动查找并恢复所有与指定插件相关的配置文件。

## 常见问题

### 使用示例

在 Windows 上设置配置：
```bash
WtfBackup.exe config -wtf "C:\Games\World of Warcraft\_retail_\WTF" -backup "D:\WoW_Backups"
WtfBackup.exe config -add-addons "DBM-Core,ElvUI,WeakAuras"
```

在 Windows 上备份魔兽世界正式服 WTF 文件夹：
```bash
WtfBackup.exe backup
```

在 Windows 上恢复所有配置的插件：
```bash
WtfBackup.exe restore
```

或单独恢复 DBM 插件的配置：
```bash
WtfBackup.exe restore -addon "DBM-Core"
```

## 注意事项

- 本程序不会压缩备份文件，以保持最高的兼容性和易用性
- 恢复时默认使用最新的备份
- 恢复插件配置时，程序会自动创建需要的文件夹结构
- 命令行参数会临时覆盖配置文件中的设置，并更新配置文件
- 程序会自动处理路径格式，支持Windows风格的反斜杠路径和Linux/macOS风格的正斜杠路径
- 在Windows上，驱动器盘符会自动规范化为大写（例如："c:" -> "C:"） 