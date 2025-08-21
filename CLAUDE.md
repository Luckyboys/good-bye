# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个生存确认服务项目，在用户一段时间没有刷新存活状态后，通过邮件发送预设的遗书内容。项目采用Go语言开发，使用每日签到方式来刷新存活状态。

**重要**: 项目已完全移除数据库依赖，现在使用配置文件和内存存储。

## 技术栈

- **后端**: Go语言 (主要开发语言)
- **前端**: HTML/CSS/JavaScript (模板引擎)
- **配置管理**: Viper + YAML配置文件
- **状态存储**: 内存存储 (签到时间)
- **邮件**: SMTP协议
- **构建工具**: Makefile
- **日志**: Logrus (结构化日志)

## 开发命令

### 构建和运行 (使用 Makefile)
```bash
# 构建调试版本 (Claude Code 测试编译用)
make debug

# 构建发行版本
make release

# 构建所有版本
make all

# 构建并运行调试版本
make run

# 清理构建文件
make clean

# 深度清理 (包括所有生成文件)
make deep-clean
```

### 测试
```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

### 代码质量
```bash
# 格式化代码
make fmt

# 运行代码检查
make lint

# 格式化并检查代码
make check

# 管理依赖
make deps
```

### 项目信息
```bash
# 显示项目信息
make info

# 显示帮助信息
make help
```

## 构建规则 (重要!)

### Claude Code 编译规则
当 Claude Code 需要编译二进制文件时：

1. **测试编译**: 使用 `make debug` 或 `go build -o builds/debug/good-bye cmd/main.go`
2. **发行编译**: 使用 `make release` 或 `go build -ldflags="-s -w" -o builds/release/good-bye cmd/main.go`
3. **永远不要** 将二进制文件输出到项目根目录
4. **确保** `logs/` 目录存在并且用于日志输出
5. **优先使用 Makefile** 进行构建

### 输出目录
- **调试版本**: `builds/debug/good-bye` (Unix) 或 `builds/debug/good-bye.exe` (Windows)
- **发行版本**: `builds/release/good-bye` (Unix) 或 `builds/release/good-bye.exe` (Windows)
- **日志文件**: `logs/` 目录
- **数据文件**: `data/` 目录

## 核心架构

### 主要组件架构

项目采用分层架构设计，包含以下核心组件：

1. **Web服务层** (`src/web/`) - 提供HTTP服务和API接口
2. **API处理层** (`src/api/`) - 处理HTTP请求和响应
3. **业务逻辑层** (`src/state/`) - 状态管理和业务逻辑
4. **服务层** (`src/email/`, `src/config/`) - 邮件和配置服务
5. **配置管理** (`src/config/`) - 配置文件管理和验证

### 关键设计模式

- **依赖注入**: 使用依赖注入管理服务之间的依赖关系
- **中间件模式**: Web请求处理使用中间件进行CORS、日志等横切关注点
- **配置驱动**: 所有配置通过YAML文件管理，支持热重载
- **内存存储**: 签到时间存储在内存中，使用读写锁保证线程安全

### 数据流架构

```
用户请求 -> Web路由 -> API处理器 -> 业务逻辑 -> 内存状态
    ↓
定时任务 -> 状态检查 -> 邮件发送 -> 配置文件读取
```

## 重要配置

### 环境变量
- `CONFIG_PATH`: 配置文件路径
- `PORT`: 服务端口 (默认: 8080)
- `LOG_LEVEL`: 日志级别

### 配置文件结构
配置文件位于 `config/config.yaml`，包含：
- 服务器配置 (端口、主机)
- 系统设置 (检查间隔、最大不活跃天数)
- 邮件服务配置 (SMTP设置)
- 部署配置 (数据目录、日志目录、遗书文件路径)

### 关键配置项
```yaml
system:
  check_interval: 24        # 检查间隔（小时）
  max_inactive_days: 7      # 最大不活跃天数
  timezone: "Asia/Shanghai"  # 时区

email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: ""
  password: ""
  from_email: ""
  test_email: ""

deployment:
  data_dir: "./data"
  log_dir: "./logs"
  posthumous_papers_file: "./data/posthumous_papers.md"
```

## 开发者角色分工

### frontend-developer-agent
负责Web界面和用户体验，使用HTML模板引擎开发前端页面。

### backend-developer-agent  
负责后端API和业务逻辑，使用Go语言开发核心功能。

### fullstack-lead-agent
负责整体架构设计和技术选型，协调各模块开发。

## 关键开发注意事项

1. **无数据库架构**: 项目已移除所有数据库依赖，使用内存存储和配置文件
2. **状态管理**: 签到时间存储在内存中，使用 `sync.RWMutex` 保证线程安全
3. **配置管理**: 所有配置通过YAML文件管理，支持热重载
4. **邮件服务**: 邮件配置从配置文件读取，不再存储在数据库中
5. **构建规则**: 严格遵守构建规则，二进制文件必须输出到 `builds/` 目录
6. **日志管理**: 日志文件必须输出到 `logs/` 目录
7. **错误处理**: 实现统一的错误处理和响应格式

## API 接口

### 健康检查
- `GET /health` - 服务健康状态

### 状态管理
- `POST /api/v1/checkin` - 用户签到
- `GET /api/v1/status` - 获取状态信息
- `GET /api/v1/stats` - 获取统计信息

### 系统设置
- `GET /api/v1/settings` - 获取系统设置
- `PUT /api/v1/settings` - 更新系统设置

### 遗书管理
- `GET /api/v1/wills` - 获取遗书状态

### 邮件服务
- `POST /api/v1/email/test` - 发送测试邮件
- `GET /api/v1/email/config` - 获取邮件配置
- `PUT /api/v1/email/config` - 更新邮件配置
- `POST /api/v1/email/config/test` - 测试邮件配置

## 项目状态

项目已完成核心功能开发，主要特点：

✅ **已实现功能**:
- 完整的Web API服务
- 内存状态管理
- 配置文件驱动的邮件服务
- 定时任务和状态检查
- 构建系统 (Makefile)
- 跨平台支持

✅ **架构优化**:
- 移除数据库依赖
- 简化系统架构
- 提高运行效率
- 减少维护复杂度

✅ **开发工具**:
- 完整的Makefile构建系统
- 代码质量检查工具
- 统一的构建规则
- 跨平台兼容性

## 提交信息规范

项目遵循 commitizen 规范的提交信息格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 提交类型 (Type)
- `feat`: 新功能
- `fix`: 修复 bug
- `chore`: 构建过程或辅助工具的变动
- `docs`: 文档更新
- `style`: 代码格式调整（不影响代码运行）
- `refactor`: 重构（既不是新增功能，也不是修改 bug）
- `perf`: 性能优化
- `test`: 增加测试

### 示例
```
feat(scheduler): 实现任务调度器重构

- 将后台任务管理重构为专门的调度器模块
- 改进退出机制，确保所有后台任务能正确响应退出信号

Refs: #123
```

## 常见问题

### Q: 为什么项目没有数据库？
A: 项目已优化为无数据库架构，所有状态存储在内存中，配置通过文件管理，提高了性能和简化了部署。

### Q: 如何编译项目？
A: 使用 `make debug` 进行测试编译，`make release` 进行发行编译。永远不要将二进制文件输出到项目根目录。

### Q: 日志文件存储在哪里？
A: 所有日志文件必须存储在 `logs/` 目录中，这是项目规则。

### Q: 如何管理配置？
A: 所有配置通过 `config/config.yaml` 文件管理，支持热重载和环境变量覆盖。

## Docker 部署

项目支持 Docker 容器化部署，提供了完整的 CI/CD 流程。

### 构建镜像
```bash
# 构建 Docker 镜像
docker build -t good-bye .

# 运行容器
docker run -d -p 8080:8080 good-bye
```

### 配置挂载
```bash
# 挂载配置文件和数据目录
docker run -d -p 8080:8080 \
  -v ./config:/root/config \
  -v ./data:/root/data \
  -v ./logs:/root/logs \
  good-bye
```

### GitHub Actions CI/CD
项目配置了完整的 GitHub Actions 工作流，包括：
- 代码质量检查（lint、vet、fmt）
- 跨平台二进制编译（Linux、Windows、macOS）
- Docker 镜像构建和推送