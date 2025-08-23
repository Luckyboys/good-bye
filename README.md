# 生存确认服务

一个在用户长时间未活动后，通过邮件发送预设遗书内容的Go项目。

## 项目概述

生存确认服务是一个简单的生命状态监控系统，用户通过定期签到来刷新存活状态。如果在指定时间内没有签到，系统会自动向预设的邮箱发送遗书内容。

### 背景说明

由于目前科技水平和厂商开放API水平，通过穿戴式设备刷新存活状态的方案暂时不太可取，没有可以调用的接口去获知或者穿戴式设备没有开放通知接口。因此使用每日签到的方式去刷新存活状态属于比较容易实现的方式。

## 功能特性

### 核心功能
- ✅ 通过命令行启动服务
- ✅ 可以配置遗书内容和标题
- ✅ 可以通过网页签到（刷新存活状态）
- ✅ 配置发送的目标邮件和发送邮箱的账号密码
- ✅ 定时检查，如果指定天数内没有刷新存活状态，则发送通知邮件
- ✅ 可以通过网页手动发送测试信件，测试收件邮箱单独配置
- ✅ 启动时，默认刷新存活状态（如果在启动命令中指定不刷新，则不刷新）
- ✅ **新增**：测试遗书发送功能，支持向第一个收件人发送测试遗书
- ✅ **新增**：邮件重试机制，支持指数退避重试（1秒到1周）
- ✅ **新增**：遗书发送后自动停止生存检测，避免重复发送

### 技术特性
- ✅ 配置文件热重载
- ✅ 状态持久化
- ✅ 日志记录和监控
- ✅ 健康检查接口
- ✅ 无数据库架构，使用内存存储和配置文件
- ✅ **新增**：现代化代码优化，使用 Go 1.18+ 最佳实践
- ✅ **新增**：Pre-commit hooks 集成，包含现代化检查
- ✅ **新增**：增强安全性，TLS 1.2+ 和文件权限优化
- ✅ **新增**：改进错误处理，支持错误包装

## 快速开始

### 系统要求
- Go 1.18+ (推荐 1.21+)
- 操作系统：Linux/macOS/Windows
- 内存：至少512MB
- 磁盘：至少100MB

### 安装步骤

1. **克隆项目**
```bash
git clone https://github.com/Luckyboys/good-bye.git
cd good-bye
```

2. **编译项目**
```bash
# 使用Makefile（推荐）
make build

# 或者直接使用go build
go build -o good-bye cmd/main.go
```

3. **配置文件**
```bash
# 复制配置文件模板
cp config/config.yaml config/config.local.yaml

# 编辑配置文件，修改邮件设置等
vim config/config.local.yaml
```

4. **创建必要目录和文件**
```bash
# 创建数据目录
mkdir -p data logs

# 创建遗书文件
echo "# 我的遗书

亲爱的家人和朋友：

如果你们收到这封信，说明我已经有一段时间没有活动了。

请照顾好自己，记住美好的时光。

爱你们的，
[你的名字]" > data/posthumous_papers.md
```

5. **启动服务**
```bash
# 基本启动
./good-bye

# 指定配置文件
./good-bye -config ./config/config.local.yaml

# 指定端口
./good-bye -port 8080
```

6. **访问服务**
打开浏览器访问：`http://localhost:8080`

## 使用方法

### Web界面使用

1. **签到功能**
   - 访问首页，点击"立即签到"按钮
   - 系统会记录签到时间并更新状态

2. **状态查看**
   - 查看当前状态（活跃/不活跃）
   - 查看最后签到时间
   - 查看不活跃时长

3. **邮件测试**
   - 点击"发送测试邮件"按钮
   - 系统会向配置的测试邮箱发送测试邮件

### 命令行使用

```bash
# 查看帮助
./good-bye --help

# 查看版本
./good-bye --version

# 启动服务
./good-bye

# 指定配置文件
./good-bye --config /path/to/config.yaml

# 指定端口
./good-bye --port 8080
```

### 配置说明

主要配置项：

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080

# 系统设置
system:
  check_interval: 24        # 检查间隔（小时）
  max_inactive_days: 7      # 最大不活跃天数
  timezone: "Asia/Shanghai"  # 时区

# 邮件配置
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  from_email: "your-email@gmail.com"
  test_email: "test@example.com"

# 部署配置
deployment:
  data_dir: "./data"
  log_dir: "./logs"
  posthumous_papers_file: "./data/posthumous_papers.md"
```

## API接口

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
- `POST /api/v1/email/test-will` - 发送测试遗书（新功能）
- `GET /api/v1/email/config` - 获取邮件配置
- `PUT /api/v1/email/config` - 更新邮件配置

## 开发和构建

### 开发环境设置
```bash
# 安装依赖
make deps

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 设置 pre-commit hooks
./scripts/setup-pre-commit.sh

# 构建调试版本
make debug

# 构建发行版本
make release
```

### Docker部署
```bash
# 根据你的情况修改遗书内容
# vim data/posthumous_papers.md

# 使用 Docker Hub 镜像
docker run -d -p 8080:8080 -v data:/root/data luckyboys/good-bye

# 或使用阿里云容器镜像
docker run -d -p 8080:8080 -v data:/root/data crpi-kyazw4facu8wslpn.cn-shanghai.personal.cr.aliyuncs.com/luckyboys/good-bye
```

## 文档

- [项目架构](docs/ARCHITECTURE.md) - 详细的技术架构说明
- [开发指南](docs/DEVELOPMENT.md) - 开发团队规划和分工
- [部署指南](docs/DEPLOYMENT.md) - 完整的部署方案和运维指南
- [构建规则](BUILD_RULES.md) - 构建系统和规则说明
- [Pre-commit Hook 设置](docs/PRE_COMMIT_HOOK.md) - 代码质量检查和现代化工具配置

## 安全注意事项

⚠️ **重要提醒**：

1. **配置文件安全**
   - `config/config.local.yaml` 包含敏感信息，已添加到 `.gitignore`
   - 请勿将包含密码的配置文件提交到版本控制
   - 建议设置适当的文件权限：`chmod 600 config/config.local.yaml`

2. **邮件安全**
   - 使用应用专用密码，不要使用主密码
   - 定期更新邮件密码
   - 确保测试邮箱的安全性

3. **网络安全**
   - 建议在生产环境中使用HTTPS
   - 配置适当的防火墙规则
   - 限制访问IP地址

## 故障排除

### 常见问题

1. **端口占用**
   ```bash
   # 查看端口占用
   netstat -tulpn | grep :8080
   # 修改配置文件中的端口
   ```

2. **邮件发送失败**
   ```bash
   # 检查邮件配置
   # 查看日志文件
   tail -f logs/good-bye.log
   # 测试邮件配置
   curl -X POST http://localhost:8080/api/v1/email/test
   ```

3. **配置文件错误**
   ```bash
   # 验证配置文件
   ./good-bye -validate-config
   ```

### 日志查看
```bash
# 查看最新日志
tail -f logs/good-bye.log

# 搜索错误
grep "ERROR" logs/good-bye.log
```

## 贡献

欢迎提交Issue和Pull Request！

### 开发流程
1. Fork项目
2. 创建功能分支
3. 提交变更
4. 创建Pull Request

### 代码规范
- 遵循Go语言标准代码风格
- 提交前运行 `make check` 进行代码检查
- 使用 pre-commit hooks 进行自动化代码质量检查
- 遵循现代化 Go 开发最佳实践
- 编写清晰的提交信息

## 许可证

本项目采用MIT许可证，详情请参阅 [LICENSE](LICENSE) 文件。

## 更新日志

### v1.1.0 (2025-08-23)
- 🚀 **新增测试遗书发送功能** - 支持向第一个收件人发送测试遗书
- 🔄 **邮件重试机制** - 实现指数退避重试，支持1秒到1周的重试间隔
- 🛡️ **生存检测停止** - 遗书发送后自动停止生存检测，避免重复发送
- 🔧 **代码现代化优化** - 使用 Go 1.18+ 最佳实践，interface{} 替换为 any
- 🔒 **安全性增强** - TLS 1.2+ 支持，文件权限优化
- ✅ **Pre-commit hooks** - 集成现代化检查工具，确保代码质量
- 🎯 **错误处理改进** - 使用 errors.As 支持错误包装
- 📚 **文档完善** - 新增 pre-commit hook 设置文档

### v1.0.0
- 初始版本发布
- 基本的签到和邮件发送功能
- Web界面和API接口
- Docker支持和CI/CD流程

## 联系方式

如有问题或建议，请通过以下方式联系：

- GitHub Issues: [提交问题](https://github.com/Luckyboys/good-bye/issues)

---

**注意**：这是一个生命状态监控工具，请谨慎使用并定期测试，确保在紧急情况下能够正常工作。