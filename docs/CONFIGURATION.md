# 配置说明

本文档详细说明了生存确认服务的配置选项、使用方法和最佳实践。

## 目录

- [配置文件结构](#配置文件结构)
- [服务器配置](#服务器配置)
- [系统配置](#系统配置)
- [日志配置](#日志配置)
- [邮件配置](#邮件配置)
- [部署配置](#部署配置)
- [时间格式说明](#时间格式说明)
- [配置示例](#配置示例)
- [环境变量](#环境变量)
- [配置验证](#配置验证)
- [常见问题](#常见问题)

## 配置文件结构

配置文件使用YAML格式，位于 `config/config.yaml`。配置文件结构如下：

```yaml
# 服务器配置
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30
  write_timeout: 30

# 系统配置
system:
  check_interval: "24h"
  max_inactive_time: "7d"
  reminder_time: "180m"
  timezone: "Asia/Shanghai"

# 日志配置
log:
  level: "info"
  format: "text"
  output: "stdout"

# 邮件配置
email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: ""
  password: ""
  from_email: ""
  test_email: "test@example.com"
  recipients:
    - email: "test@example.com"
      name: "测试用户"

# 部署配置
deployment:
  posthumous_papers_file: "./data/posthumous_papers.md"
```

## 服务器配置

### `server.port`
- **类型**: `int`
- **默认值**: `8080`
- **说明**: HTTP服务器监听端口
- **示例**: `8080`, `3000`, `80`

### `server.host`
- **类型**: `string`
- **默认值**: `"0.0.0.0"`
- **说明**: HTTP服务器监听地址
- **示例**: `"0.0.0.0"`, `"localhost"`, `"127.0.0.1"`

### `server.read_timeout`
- **类型**: `int`
- **默认值**: `30`
- **说明**: 读取超时时间（秒）
- **示例**: `30`, `60`, `120`

### `server.write_timeout`
- **类型**: `int`
- **默认值**: `30`
- **说明**: 写入超时时间（秒）
- **示例**: `30`, `60`, `120`

## 系统配置

### `system.check_interval`
- **类型**: `string` (时间格式)
- **默认值**: `"24h"`
- **说明**: 系统检查间隔时间
- **示例**: `"1h"`, `"30m"`, `"24h"`, `"7d"`

### `system.max_inactive_time`
- **类型**: `string` (时间格式)
- **默认值**: `"7d"`
- **说明**: 最大不活跃时间，超过此时间将发送遗书
- **示例**: `"1d"`, `"3d"`, `"7d"`, `"30d"`

### `system.reminder_time`
- **类型**: `string` (时间格式)
- **默认值**: `"180m"`
- **说明**: 发送遗书前的提醒时间，设置为`0`则不发送提醒
- **示例**: `"0"`, `"30m"`, `"1h"`, `"3h"`

### `system.timezone`
- **类型**: `string`
- **默认值**: `"Asia/Shanghai"`
- **说明**: 系统时区设置
- **示例**: `"Asia/Shanghai"`, `"UTC"`, `"America/New_York"`

## 日志配置

### `log.level`
- **类型**: `string`
- **默认值**: `"info"`
- **说明**: 日志级别
- **可选值**: `"debug"`, `"info"`, `"warn"`, `"error"`, `"fatal"`

### `log.format`
- **类型**: `string`
- **默认值**: `"text"`
- **说明**: 日志格式
- **可选值**: `"text"`, `"json"`

### `log.output`
- **类型**: `string`
- **默认值**: `"stdout"`
- **说明**: 日志输出位置
- **示例**: `"stdout"`, `"logs/good-bye.log"`

## 邮件配置

### `email.smtp_host`
- **类型**: `string`
- **默认值**: `"smtp.gmail.com"`
- **说明**: SMTP服务器地址
- **示例**: `"smtp.gmail.com"`, `"smtp.qq.com"`, `"smtp.163.com"`

### `email.smtp_port`
- **类型**: `int`
- **默认值**: `587`
- **说明**: SMTP服务器端口
- **示例**: `587` (TLS), `465` (SSL)

### `email.username`
- **类型**: `string`
- **默认值**: `""`
- **说明**: SMTP用户名
- **示例**: `"your-email@gmail.com"`

### `email.password`
- **类型**: `string`
- **默认值**: `""`
- **说明**: SMTP密码或应用专用密码
- **示例**: `"your-app-password"`

### `email.from_email`
- **类型**: `string`
- **默认值**: `""`
- **说明**: 发件人邮箱地址
- **示例**: `"your-email@gmail.com"`

### `email.test_email`
- **类型**: `string`
- **默认值**: `"test@example.com"`
- **说明**: 测试邮件接收地址
- **示例**: `"test@example.com"`

### `email.recipients`
- **类型**: `array`
- **默认值**: `[]`
- **说明**: 遗书邮件接收人列表
- **结构**:
  ```yaml
  recipients:
    - email: "recipient1@example.com"
      name: "接收人1"
    - email: "recipient2@example.com"
      name: "接收人2"
  ```

## 部署配置

### `deployment.posthumous_papers_file`
- **类型**: `string`
- **默认值**: `"./data/posthumous_papers.md"`
- **说明**: 遗书文件路径
- **示例**: `"./data/posthumous_papers.md"`, `"/path/to/will.md"`

## 时间格式说明

系统支持以下时间格式：

- **秒**: `30s` (30秒)
- **分钟**: `5m` (5分钟), `180m` (180分钟)
- **小时**: `1h` (1小时), `24h` (24小时)
- **天**: `1d` (1天), `7d` (7天)
- **周**: `1w` (1周), `4w` (4周)

## 配置示例

### 基础配置示例
```yaml
server:
  port: 8080
  host: "0.0.0.0"

system:
  check_interval: "24h"
  max_inactive_time: "7d"
  reminder_time: "3h"
  timezone: "Asia/Shanghai"

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

### 生产环境配置示例
```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 60
  write_timeout: 60

system:
  check_interval: "12h"
  max_inactive_time: "3d"
  reminder_time: "6h"
  timezone: "Asia/Shanghai"

log:
  level: "info"
  format: "json"
  output: "logs/good-bye.log"

email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "service@company.com"
  password: "app-specific-password"
  from_email: "service@company.com"
  test_email: "admin@company.com"
  recipients:
    - email: "emergency@company.com"
      name: "应急联系人"
    - email: "family@personal.com"
      name: "家人"

deployment:
  posthumous_papers_file: "/opt/good-bye/data/posthumous_papers.md"
```

### 开发环境配置示例
```yaml
server:
  port: 3000
  host: "localhost"

system:
  check_interval: "1m"
  max_inactive_time: "5m"
  reminder_time: "1m"
  timezone: "Asia/Shanghai"

log:
  level: "debug"
  format: "text"
  output: "stdout"

email:
  smtp_host: "smtp.mailtrap.io"
  smtp_port: 2525
  username: "mailtrap-user"
  password: "mailtrap-pass"
  from_email: "test@example.com"
  test_email: "dev@example.com"
  recipients:
    - email: "dev-test@example.com"
      name: "开发测试"

deployment:
  posthumous_papers_file: "./data/test_posthumous_papers.md"
```

## 环境变量

可以通过环境变量覆盖配置文件中的设置：

- `CONFIG_PATH`: 配置文件路径
- `PORT`: 服务器端口（覆盖 `server.port`）
- `HOST`: 服务器地址（覆盖 `server.host`）
- `LOG_LEVEL`: 日志级别（覆盖 `log.level`）
- `TIMEZONE`: 时区（覆盖 `system.timezone`）

### 环境变量使用示例
```bash
# 使用环境变量覆盖配置
export PORT=9090
export LOG_LEVEL=debug
export CONFIG_PATH=/path/to/config.yaml

./good-bye
```

## 配置验证

### 命令行验证
```bash
# 验证配置文件
./good-bye -validate-config

# 指定配置文件验证
./good-bye -config ./config/local.yaml -validate-config
```

### 验证内容
配置验证会检查以下内容：

1. **配置文件格式**: YAML格式是否正确
2. **必需字段**: 是否缺少必需的配置项
3. **数据类型**: 配置值的数据类型是否正确
4. **端口范围**: 端口号是否在有效范围内 (1-65535)
5. **时间格式**: 时间格式是否正确
6. **时区**: 时区是否有效
7. **文件路径**: 遗书文件路径是否存在
8. **邮件配置**: SMTP配置是否完整

## 常见问题

### Q: 如何配置Gmail SMTP？
A: 
1. 启用两步验证
2. 生成应用专用密码
3. 使用应用专用密码作为 `email.password`
4. 配置示例：
   ```yaml
   email:
     smtp_host: "smtp.gmail.com"
     smtp_port: 587
     username: "your-email@gmail.com"
     password: "your-app-password"
   ```

### Q: 如何配置QQ邮箱？
A:
```yaml
email:
  smtp_host: "smtp.qq.com"
  smtp_port: 587
  username: "your-email@qq.com"
  password: "your-auth-code"
```

### Q: 提醒时间如何设置？
A: 在 `system.reminder_time` 中设置时间，设置为 `0` 则不发送提醒：
```yaml
system:
  reminder_time: "3h"  # 提前3小时发送提醒
  reminder_time: "0"   # 不发送提醒
```

### Q: 如何测试邮件配置？
A: 
1. 使用Web界面的"发送测试邮件"功能
2. 使用API: `POST /api/v1/email/test`
3. 使用命令行工具测试

### Q: 如何配置多个收件人？
A: 在 `email.recipients` 中添加多个收件人：
```yaml
email:
  recipients:
    - email: "person1@example.com"
      name: "收件人1"
    - email: "person2@example.com"
      name: "收件人2"
```

### Q: 如何调整检查频率？
A: 修改 `system.check_interval`：
```yaml
system:
  check_interval: "12h"  # 每12小时检查一次
  check_interval: "1h"   # 每1小时检查一次
```

### Q: 如何配置生产环境日志？
A: 修改日志配置为文件输出和JSON格式：
```yaml
log:
  level: "info"
  format: "json"
  output: "logs/good-bye.log"
```

### Q: 如何使用不同的时区？
A: 修改 `system.timezone`：
```yaml
system:
  timezone: "UTC"           # UTC时间
  timezone: "America/New_York"  # 纽约时间
  timezone: "Asia/Tokyo"    # 东京时间
```

## 最佳实践

1. **安全性**: 
   - 不要将包含密码的配置文件提交到版本控制
   - 使用应用专用密码而不是主密码
   - 设置适当的文件权限：`chmod 600 config/config.yaml`

2. **配置管理**:
   - 为不同环境创建不同的配置文件
   - 使用环境变量覆盖敏感信息
   - 定期备份配置文件

3. **测试验证**:
   - 在生产环境部署前先测试配置
   - 使用 `-validate-config` 验证配置文件
   - 测试邮件发送功能

4. **监控**:
   - 监控日志文件以发现配置问题
   - 定期检查系统状态
   - 设置适当的日志级别
