# 部署指南

## 概述

本文档提供了生存确认服务的完整部署方案，包括Docker部署、CI/CD流程和手动部署方式。

## 部署方式

### 1. Docker部署（推荐）

#### 构建镜像
```bash
# 使用Makefile构建（推荐）
make docker-build

# 这将构建带有版本号+日期标签的Docker镜像
# 例如：good-bye:v1.2.0-20250908

# 查看构建的镜像
docker images | grep good-bye
```

镜像标签格式：
- **完整标签**: `v1.2.0-20250908` (版本号+年月日)
- **版本标签**: `v1.2.0` (仅版本号)
- **最新标签**: `latest` (最新版本)

#### 运行容器
```bash
# 使用Makefile运行（推荐）
make docker-run

# 基本运行
docker run -d -p 8080:8080 good-bye:v1.2.0-20250908

# 挂载配置文件和数据目录
docker run -d \
  -p 8080:8080 \
  -v ./config:/root/config \
  -v ./data:/root/data \
  -v ./logs:/root/logs \
  good-bye:v1.2.0-20250908

# 使用环境变量配置
docker run -d \
  -p 8080:8080 \
  -e CONFIG_PATH=/root/config/config.yaml \
  -e PORT=8080 \
  -e LOG_LEVEL=info \
  good-bye:v1.2.0-20250908

# 使用特定版本的镜像
docker run -d -p 8080:8080 good-bye:v1.2.0
```

#### Docker Compose部署
创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  good-bye:
    build: 
      context: .
      dockerfile: Dockerfile
    image: good-bye:v1.2.0-20250908
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

### 2. 手动部署

#### 系统要求
- Go 1.24+ (已测试版本)
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
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30
  write_timeout: 30

system:
  check_interval: "24h"
  max_inactive_time: "7d"
  reminder_time: "3h"
  timezone: "Asia/Shanghai"

log:
  level: "info"
  format: "text"
  output: "stdout"

email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  from_email: "your-email@gmail.com"
  test_email: "test@example.com"
  recipients:
    - email: "family@example.com"
      name: "家人"
    - email: "friend@example.com"
      name: "朋友"

deployment:
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

## 重要信息

### 银行信息
- **银行名称**: XXX银行
- **账号**: ***********
- **密码**: 请联系律师获取

### 数字资产
- **比特币钱包**: 恢复短语安全保存
- **社交媒体**: 账号已设置自动删除
- **云存储**: 文件已加密备份

## 最后的话

请照顾好自己，记住我们一起度过的美好时光。不要为我的离开而悲伤，这是自然规律。

## 联系方式
如果需要帮助，请联系：
- **律师**: 张律师，电话：138-XXXX-XXXX
- **保险代理人**: 李经理，电话：139-XXXX-XXXX

爱你们的，
[你的名字]

---
*这封信使用Markdown格式编写，会自动转换为HTML邮件发送*"
> data/posthumous_papers.md
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
   - Docker镜像推送到多个容器镜像仓库，使用版本号+年月日标签
   - 推送到 Docker Hub：`${{ secrets.DOCKER_USERNAME }}/good-bye`
   - 推送到阿里云容器镜像服务：`crpi-kyazw4facu8wslpn.cn-shanghai.personal.cr.aliyuncs.com/luckyboys/good-bye`
   - 镜像标签格式：`v1.2.0-20250908`（版本号+年月日）
   - 自动化部署到生产环境

### 工作流文件
`.github/workflows/ci-cd.yml` 包含完整的CI/CD流程配置。

### 所需 GitHub Secrets
需要在 GitHub 仓库设置中配置以下 secrets：
- `DOCKER_USERNAME`: Docker Hub 用户名
- `DOCKER_PASSWORD`: Docker Hub 密码
- `ALIYUN_REGISTRY_USERNAME`: 阿里云容器镜像服务用户名
- `ALIYUN_REGISTRY_PASSWORD`: 阿里云容器镜像服务密码

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
- 提醒状态
- 重试次数

### 新功能监控
- **提醒系统**: 监控提醒邮件的发送状态
- **重试机制**: 监控邮件重试次数和成功率
- **Markdown转换**: 监控遗书文件转换状态
- **生存检测停止**: 监控遗书发送后的系统状态

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
- 测试邮件发送功能
- 验证提醒系统工作状态

## 镜像标签策略

项目使用版本号+年月日的镜像标签策略：

### 标签格式
- **完整标签**: `v1.2.0-20250908` (版本号+年月日)
- **版本标签**: `v1.2.0` (仅版本号)
- **最新标签**: `latest` (最新版本)

### 标签优势
- **版本追踪**: 清晰标识每个构建的版本和日期
- **回滚支持**: 可以轻松回滚到特定版本
- **多环境部署**: 不同环境可以使用不同版本的镜像
- **自动化**: CI/CD流水线自动生成和管理标签

### 使用示例
```bash
# 拉取特定版本的镜像
docker pull username/good-bye:v1.2.0-20250908

# 拉取最新版本
docker pull username/good-bye:latest

# 拉取版本号（指向最新的该版本构建）
docker pull username/good-bye:v1.2.0
```

## 高级部署选项

### 1. Kubernetes部署
```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: good-bye
spec:
  replicas: 2
  selector:
    matchLabels:
      app: good-bye
  template:
    metadata:
      labels:
        app: good-bye
    spec:
      containers:
      - name: good-bye
        image: luckyboys/good-bye:latest
        ports:
        - containerPort: 8080
        env:
        - name: CONFIG_PATH
          value: "/app/config/config.yaml"
        volumeMounts:
        - name: config
          mountPath: "/app/config"
        - name: data
          mountPath: "/app/data"
        - name: logs
          mountPath: "/app/logs"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: good-bye-config
      - name: data
        persistentVolumeClaim:
          claimName: good-bye-data
      - name: logs
        persistentVolumeClaim:
          claimName: good-bye-logs
```

### 2. 反向代理配置 (Nginx)
```nginx
# nginx.conf
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # 静态文件缓存
    location /static/ {
        proxy_pass http://localhost:8080;
        expires 1d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 3. 监控集成 (Prometheus + Grafana)
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'good-bye'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### 4. 自动化脚本
```bash
#!/bin/bash
# deploy.sh - 自动化部署脚本

set -e

# 停止现有服务
sudo systemctl stop good-bye || true

# 备份
cp -r /opt/good-bye /opt/good-bye.backup.$(date +%Y%m%d_%H%M%S)

# 更新代码
cd /opt/good-bye
git pull origin main

# 构建
make release

# 更新配置
cp config/config.local.yaml /opt/good-bye/config/

# 启动服务
sudo systemctl start good-bye

# 健康检查
sleep 5
if curl -f http://localhost:8080/health; then
    echo "部署成功！"
else
    echo "部署失败，正在回滚..."
    sudo systemctl stop good-bye
    cp -r /opt/good-bye.backup.$(date +%Y%m%d_%H%M%S)/* /opt/good-bye/
    sudo systemctl start good-bye
    exit 1
fi
```