# 开发指南

## 概述

本文档提供了生存确认服务项目的开发指南，包括开发环境设置、代码规范、测试策略和最佳实践。

## 开发环境设置

### 前置要求
- Go 1.24.5 或更高版本
- Git
- Make (可选，用于构建)
- Docker (可选，用于容器化)

### 环境设置步骤

1. **克隆项目**
   ```bash
   git clone https://github.com/Luckyboys/good-bye.git
   cd good-bye
   ```

2. **安装Go依赖**
   ```bash
   go mod tidy
   ```

3. **创建配置文件**
   ```bash
   cp config/config.yaml config/config.local.yaml
   # 编辑配置文件
   vim config/config.local.yaml
   ```

4. **创建必要目录**
   ```bash
   mkdir -p data logs
   ```

5. **设置Pre-commit Hooks**
   ```bash
   ./scripts/setup-pre-commit.sh
   ```

### 开发工具配置

#### IDE配置
- **VSCode**: 推荐安装Go插件
- **GoLand**: 配置Go模块支持
- **Vim**: 配置vim-go插件

#### 调试配置
- 使用Delve调试器
- 配置launch.json进行VSCode调试
- 使用GoLand内置调试器

## 代码规范

### Go代码规范
- 遵循Go官方代码规范
- 使用gofmt格式化代码
- 使用golint进行代码检查
- 使用go vet进行静态检查

### 提交信息规范
遵循Conventional Commits规范：
```
<type>(<scope>): <subject>

<body>

<footer>
```

**提交类型**:
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建过程或辅助工具变动

### 文件命名规范
- Go文件使用小写字母和下划线
- 每个文件应该有明确的职责
- 测试文件以`_test.go`结尾

## 模块开发指南

### 1. Web服务层开发
负责路由配置和HTTP服务：

```go
// src/web/router.go
func SetupRouter() *gin.Engine {
    router := gin.Default()
    
    // 中间件
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())
    
    // 路由配置
    router.GET("/", handlers.Index)
    router.POST("/api/v1/checkin", handlers.Checkin)
    
    return router
}
```

### 2. API处理器开发
实现具体的业务逻辑：

```go
// src/api/handlers/status.go
func GetStatus(c *gin.Context) {
    status := stateManager.GetStatus()
    response.Success(c, status)
}
```

### 3. 状态管理开发
管理内存状态和线程安全：

```go
// src/state/state_manager.go
type StateManager struct {
    mu       sync.RWMutex
    lastCheckin time.Time
    isActive    bool
}

func (sm *StateManager) UpdateCheckin() {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    sm.lastCheckin = time.Now()
    sm.isActive = true
}
```

### 4. 邮件服务开发
处理邮件发送和重试：

```go
// src/email/email_service.go
type EmailService struct {
    smtpHost string
    smtpPort int
    retryManager *RetryManager
}

func (es *EmailService) SendEmail(to, subject, body string) error {
    // 邮件发送逻辑
    return es.retryManager.Retry(func() error {
        return es.sendWithRetry(to, subject, body)
    })
}
```

## 测试策略

### 单元测试
- 每个Go文件都应该有对应的测试文件
- 使用Go内置的testing包
- 测试覆盖率应该达到80%以上

### 集成测试
- 测试API接口
- 测试邮件发送功能
- 测试状态管理

### 端到端测试
- 测试完整的用户流程
- 使用Docker进行容器化测试

### 测试命令
```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 运行特定测试
go test ./src/state -v
```

## 构建和部署

### 本地构建
```bash
# 构建调试版本
make debug

# 构建发行版本
make release

# 运行测试
make test
```

### 代码质量检查
```bash
# 格式化代码
make fmt

# 运行代码检查
make lint

# 运行所有检查
make check
```

### CI/CD流程
- GitHub Actions自动运行测试
- 代码合并时自动构建
- 发布时自动创建Docker镜像

## 调试和故障排除

### 调试技巧
- 使用logrus日志进行调试
- 使用Delve调试器进行断点调试
- 使用pprof进行性能分析

### 常见问题
1. **编译错误**: 检查Go版本和依赖
2. **运行时错误**: 查看日志文件
3. **配置错误**: 验证配置文件格式
4. **邮件发送失败**: 检查SMTP配置

### 日志查看
```bash
# 查看最新日志
tail -f logs/good-bye.log

# 搜索错误
grep "ERROR" logs/good-bye.log
```

## 性能优化

### 代码优化
- 使用sync.Pool减少内存分配
- 避免不必要的锁竞争
- 使用缓存提高性能

### 内存优化
- 及时释放不再使用的资源
- 使用对象池减少GC压力
- 避免内存泄漏

### 并发优化
- 使用goroutine处理并发请求
- 使用channel进行通信
- 避免竞态条件

## 安全考虑

### 代码安全
- 不在代码中硬编码敏感信息
- 使用环境变量存储密钥
- 定期更新依赖包

### 部署安全
- 使用HTTPS加密通信
- 配置适当的防火墙规则
- 定期更新系统和依赖

## 协作方式

### 工作流程
1. **需求分析**: 技术负责人负责需求分析和架构设计
2. **任务分配**: 根据需求将任务分配给相应的开发人员
3. **并行开发**: 各模块负责人进行开发工作
4. **代码审查**: 所有代码变更都需要经过审查
5. **集成测试**: 自动化测试和CI/CD验证

### 沟通机制
- 定期技术会议：每周进行技术进度同步
- 代码审查：所有代码变更都需要经过审查
- 文档同步：及时更新技术文档和API文档
- 问题跟踪：使用Issue跟踪系统管理问题和任务

### 分支策略
- **main**: 主分支，保持稳定可部署状态
- **feature**: 功能分支，用于开发新功能
- **hotfix**: 紧急修复分支，用于修复生产问题
- **develop**: 开发分支，用于集成测试

## 技术栈

### 核心技术栈
- **Go 1.24+**: 主要开发语言
- **Gin**: Web框架
- **Viper**: 配置管理
- **Logrus**: 日志管理
- **Cron**: 定时任务
- **HTML/CSS/JavaScript**: 前端技术
- **Gin Template**: Go模板引擎

### 开发工具
- **Delve**: Go调试器
- **gofmt/golint**: 代码格式化和检查
- **go vet**: 静态分析
- **Makefile**: 构建工具
- **Docker**: 容器化
- **GitHub Actions**: CI/CD

### 测试工具
- **Go testing**: 单元测试
- **Testify**: 测试断言
- **HTTP测试**: API测试
- **覆盖率测试**: 测试覆盖率分析

## 贡献指南

### 贡献流程
1. **Fork项目**到个人仓库
2. **创建功能分支**: `git checkout -b feature/new-feature`
3. **开发功能**: 遵循代码规范和最佳实践
4. **运行测试**: 确保所有测试通过
5. **提交代码**: 使用规范的提交信息
6. **创建Pull Request**: 提供详细的变更说明

### 代码审查标准
- 代码符合项目规范
- 功能测试覆盖完整
- 性能考虑充分
- 安全性检查通过
- 文档更新完整

### 问题报告
- 使用GitHub Issues报告问题
- 提供详细的问题描述
- 包含复现步骤和期望结果
- 提供相关日志和环境信息

## 发布流程

### 版本号规范
遵循语义化版本控制 (SemVer)：
- **主版本号**: 不兼容的API变更
- **次版本号**: 向后兼容的功能新增
- **修订版本号**: 向后兼容的问题修复

### 发布步骤
1. **更新版本号**: 在所有相关文件中更新版本号
2. **更新变更日志**: 记录所有变更
3. **创建发布分支**: `git checkout -b release/v1.2.0`
4. **测试验证**: 运行完整的测试套件
5. **创建标签**: `git tag -a v1.2.0 -m "Release v1.2.0"`
6. **发布到GitHub**: 创建GitHub Release
7. **构建Docker镜像**: 推送到Docker Hub

### 回滚策略
- 使用Git标签快速回滚
- Docker镜像版本管理
- 数据备份和恢复方案

## 监控和维护

### 应用监控
- 健康检查端点: `/health`
- 性能指标监控
- 错误日志监控
- 资源使用监控

### 日志管理
- 结构化日志输出
- 日志轮转配置
- 错误日志聚合
- 性能日志分析

### 维护任务
- 定期更新依赖包
- 安全漏洞检查
- 性能优化
- 文档更新