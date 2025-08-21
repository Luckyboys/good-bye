# 部署指南

## 概述

本文档提供了生存确认服务的完整部署方案，包括Docker部署、CI/CD流程和手动部署方式。

## 部署方式

### 1. Docker部署（推荐）

#### 构建镜像
```bash
# 构建Docker镜像
docker build -t good-bye:latest .

# 或者使用Makefile
make docker-build
```

#### 运行容器
```bash
# 基本运行
docker run -d -p 8080:8080 good-bye:latest

# 挂载配置文件和数据目录
docker run -d \
  -p 8080:8080 \
  -v ./config:/root/config \
  -v ./data:/root/data \
  -v ./logs:/root/logs \
  good-bye:latest

# 使用环境变量配置
docker run -d \
  -p 8080:8080 \
  -e CONFIG_PATH=/root/config/config.yaml \
  -e PORT=8080 \
  -e LOG_LEVEL=info \
  good-bye:latest
```

#### Docker Compose部署
创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  good-bye:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config:/root/config
      - ./data:/root/data
      - ./logs:/root/logs
    environment:
      - CONFIG_PATH=/root/config/config.yaml
      - PORT=8080
      - LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

启动服务：
```bash
docker-compose up -d
```

### 2. 手动部署

#### 系统要求
- Go 1.20+
- 操作系统：Linux/macOS/Windows
- 内存：至少512MB
- 磁盘：至少100MB

#### 编译程序
```bash
# 克隆项目
git clone https://github.com/Luckyboys/good-bye.git
cd good-bye

# 编译
make build

# 或者直接使用go build
go build -o good-bye cmd/main.go
```

#### 配置文件
1. 复制配置文件模板：
```bash
cp config/config.yaml config/config.local.yaml
```

2. 编辑配置文件：
```yaml
# config/config.local.yaml
server:
  host: "0.0.0.0"
  port: 8080

system:
  check_interval: 24
  max_inactive_days: 7
  timezone: "Asia/Shanghai"

email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "your-email@gmail.com"
  password: "your-password"
  from_email: "your-email@gmail.com"
  test_email: "test@example.com"

deployment:
  data_dir: "./data"
  log_dir: "./logs"
  posthumous_papers_file: "./data/posthumous_papers.md"
```

#### 创建必要目录
```bash
mkdir -p data logs
```

#### 创建遗书文件
```bash
echo "# 我的遗书

亲爱的家人和朋友：

如果你们收到这封信，说明我已经有一段时间没有活动了。

请照顾好自己，记住美好的时光。

爱你们的，
[你的名字]" > data/posthumous_papers.md
```

#### 启动服务
```bash
# 基本启动
./good-bye

# 指定配置文件
./good-bye -config ./config/config.local.yaml

# 指定端口
./good-bye -port 8080

# 后台运行（Linux/macOS）
nohup ./good-bye > /dev/null 2>&1 &
```

### 3. 系统服务部署（Linux）

#### 创建systemd服务文件
```bash
sudo tee /etc/systemd/system/good-bye.service > /dev/null <<EOF
[Unit]
Description=Good-Bye Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/good-bye
ExecStart=/path/to/good-bye -config /path/to/config/config.local.yaml
Restart=always
RestartSec=5
Environment=CONFIG_PATH=/path/to/config/config.local.yaml
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
EOF
```

#### 启用服务
```bash
# 重载systemd配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start good-bye

# 设置开机自启
sudo systemctl enable good-bye

# 查看服务状态
sudo systemctl status good-bye

# 查看日志
sudo journalctl -u good-bye -f
```

## CI/CD部署

### GitHub Actions自动部署

项目已配置GitHub Actions工作流，支持：

1. **代码质量检查**
   - Golint检查
   - Go vet检查
   - 格式化检查
   - 测试覆盖率

2. **自动构建**
   - 多平台二进制文件构建
   - Docker镜像构建
   - 版本标记

3. **自动部署**
   - Docker镜像推送
   - 自动化部署到生产环境

### 工作流文件
`.github/workflows/ci-cd.yml` 包含完整的CI/CD流程配置。

## 监控和日志

### 健康检查
服务提供健康检查接口：
```bash
curl http://localhost:8080/health
```

### 日志配置
日志文件位置：`./logs/`
支持配置日志级别：
- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息

### 监控指标
- 服务状态
- 签到次数
- 邮件发送状态
- 系统资源使用情况

## 安全配置

### 1. 配置文件安全
- `config.local.yaml` 包含敏感信息，不要提交到版本控制
- 设置适当的文件权限：
```bash
chmod 600 config/config.local.yaml
```

### 2. 网络安全
- 使用防火墙限制访问
- 考虑使用HTTPS
- 配置访问控制

### 3. 邮件安全
- 使用应用专用密码
- 启用SMTP认证
- 定期更新密码

## 备份和恢复

### 备份策略
1. **配置文件备份**
```bash
cp config/config.local.yaml backups/
```

2. **数据文件备份**
```bash
cp data/posthumous_papers.md backups/
```

3. **日志备份**
```bash
tar -czf logs-$(date +%Y%m%d).tar.gz logs/
```

### 恢复策略
1. 恢复配置文件
2. 恢复数据文件
3. 重启服务

## 故障排除

### 常见问题

1. **端口占用**
```bash
# 查看端口占用
netstat -tulpn | grep :8080

# 修改配置文件中的端口
```

2. **权限问题**
```bash
# 检查文件权限
ls -la config/config.local.yaml

# 修复权限
chmod 644 config/config.yaml
chmod 600 config/config.local.yaml
```

3. **邮件发送失败**
```bash
# 检查邮件配置
# 查看日志文件
tail -f logs/good-bye.log

# 测试邮件配置
curl -X POST http://localhost:8080/api/v1/email/test
```

4. **配置文件错误**
```bash
# 验证配置文件
./good-bye -validate-config
```

### 日志分析
```bash
# 查看最新日志
tail -f logs/good-bye.log

# 搜索错误
grep "ERROR" logs/good-bye.log

# 统计签到次数
grep "签到成功" logs/good-bye.log | wc -l
```

## 性能优化

### 1. 资源优化
- 调整日志级别
- 配置适当的检查间隔
- 使用连接池

### 2. 并发优化
- 增加Worker数量
- 使用缓存
- 优化数据库查询

### 3. 网络优化
- 使用CDN
- 启用压缩
- 配置keep-alive

## 更新和维护

### 版本更新
```bash
# 备份当前版本
cp good-bye good-bye.backup

# 下载新版本
git pull origin main

# 重新编译
make build

# 重启服务
sudo systemctl restart good-bye
```

### 配置更新
```bash
# 编辑配置文件
vim config/config.local.yaml

# 重启服务以应用新配置
sudo systemctl restart good-bye
```

### 定期维护
- 检查日志文件大小
- 清理过期日志
- 更新依赖包
- 检查安全更新