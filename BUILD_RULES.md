# Good-Bye Service 构建规则

## 概述
此项目遵循严格的构建规则，确保二进制文件和日志文件不会污染项目根目录。

## 目录结构

```
good-bye/
├── builds/
│   ├── debug/          # 调试版本二进制文件
│   └── release/        # 发行版本二进制文件
├── logs/               # 日志文件目录
├── data/               # 数据文件目录
├── config/             # 配置文件目录
├── src/                # 源代码
├── cmd/                # 命令行入口
├── build.sh            # Unix/Linux/macOS 构建脚本
├── build.bat           # Windows 构建脚本
└── .gitignore          # Git 忽略文件
```

## 构建规则

### 1. 二进制文件位置
- **调试版本**: `builds/debug/good-bye` (Unix) 或 `builds/debug/good-bye.exe` (Windows)
- **发行版本**: `builds/release/good-bye` (Unix) 或 `builds/release/good-bye.exe` (Windows)

### 2. 日志文件位置
- **所有日志文件**: `logs/` 目录
- **配置文件中的默认日志目录**: `./logs`

### 3. 禁止的构建位置
- ❌ 项目根目录 (例如: `./good-bye`)
- ❌ 源代码目录 (例如: `./cmd/good-bye`)
- ❌ 任何其他非 `builds/` 目录

## 构建方法

### Unix/Linux/macOS (推荐使用 Makefile)
```bash
# 构建调试版本 (默认)
make debug

# 构建发行版本
make release

# 构建两个版本
make all

# 清理构建文件
make clean

# 深度清理
make deep-clean

# 运行调试版本
make run

# 运行测试
make test

# 代码格式化
make fmt

# 显示帮助
make help
```

### Windows
在 Windows 上，Makefile 同样适用。确保系统已安装 make：

1. **安装 make**：
   - 使用 Chocolatey：`choco install make`
   - 使用 Scoop：`scoop install make`
   - 或从 [GnuWin32](http://gnuwin32.sourceforge.net/packages/make.htm) 下载

2. **使用 Makefile**：
   ```cmd
   make debug          # 构建调试版本
   make release        # 构建发行版本
   make all            # 构建两个版本
   make clean          # 清理构建文件
   make help           # 显示帮助
   ```

### 手动构建 (不推荐)
如果必须手动构建，请确保输出到正确的目录：

```bash
# 调试版本
go build -o builds/debug/good-bye cmd/main.go

# 发行版本
go build -ldflags="-s -w" -o builds/release/good-bye cmd/main.go
```

## Git 忽略规则

`.gitignore` 文件会忽略：
- 所有二进制文件 (`good-bye`, `good-bye.exe`)
- `builds/` 目录
- `logs/` 目录
- `data/` 目录
- IDE 文件
- 临时文件

## 配置文件

默认配置文件 (`config/config.yaml`) 中的日志目录设置：
```yaml
deployment:
  log_dir: "./logs"     # 日志文件目录
  data_dir: "./data"     # 数据文件目录
  backup_dir: "./backups" # 备份目录
```

## Claude Code 规则

当 Claude Code 需要编译二进制文件时：

1. **测试编译**: 使用 `make debug` 或 `go build -o builds/debug/good-bye cmd/main.go`
2. **发行编译**: 使用 `make release` 或 `go build -ldflags="-s -w" -o builds/release/good-bye cmd/main.go`
3. **永远不要** 将二进制文件输出到项目根目录
4. **确保** `logs/` 目录存在并且用于日志输出
5. **优先使用 Makefile** 进行构建，因为它提供了更多的功能和更好的错误处理

## 自动化

### Unix/Linux/macOS (推荐)
建议使用 Makefile，它提供了完整的项目管理功能：

```bash
# 基本构建
make debug          # 构建调试版本
make release        # 构建发行版本
make all            # 构建两个版本

# 运行和测试
make run            # 构建并运行调试版本
make test           # 运行测试
make test-coverage  # 运行测试并生成覆盖率报告

# 代码质量
make fmt            # 格式化代码
make lint           # 运行代码检查
make check          # 格式化并检查代码

# 清理
make clean          # 清理构建文件
make deep-clean     # 深度清理所有生成文件

# 依赖管理
make deps           # 下载依赖
make deps-graph     # 生成依赖图

# 信息和帮助
make info           # 显示项目信息
make help           # 显示帮助信息
```

Makefile 提供的优势：
- 自动创建必要的目录
- 设置正确的构建标志
- 提供版本信息
- 清理旧的构建文件
- 提供彩色输出和错误处理
- 支持测试、代码检查、依赖管理等功能
- 完全跨平台兼容（包括 Windows）