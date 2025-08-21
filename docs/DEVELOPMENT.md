# 开发团队Agent规划

## 概述

本文档描述了生存确认服务项目的开发团队规划和角色分工。

## 开发者Agent分工

### 1. frontend-developer-agent

**职责**: 前端开发工程师
- 开发Web界面和用户体验
- 实现签到页面和配置管理界面
- 处理前端交互和响应式设计
- 优化用户界面和用户体验

**负责模块**:
- `src/web/` - Web服务层
- `templates/` - HTML模板
- `static/` - 静态资源(CSS, JS, 图片)
- 前端构建和打包

**开发工具**:
- HTML/CSS/JavaScript开发
- 模板引擎集成
- 前端构建工具
- 浏览器调试工具

### 2. backend-developer-agent

**职责**: 后端开发工程师
- 设计和实现API接口
- 开发业务逻辑和数据处理
- 实现状态管理和持久化
- 处理HTTP请求和响应

**负责模块**:
- `src/api/` - API接口层
- `src/handlers/` - 请求处理器
- `src/state/` - 状态管理
- `src/models/` - 数据模型

**开发工具**:
- Go语言开发工具
- HTTP服务器开发
- 数据库操作工具
- API测试工具

### 3. devops-engineer-agent

**职责**: DevOps工程师
- 设计Docker容器化方案
- 配置CI/CD流水线
- 管理部署和运维
- 监控和日志管理

**负责模块**:
- `docker/` - Docker配置
- `.github/workflows/` - CI/CD配置
- `scripts/` - 部署脚本
- 监控和日志配置

**开发工具**:
- Docker构建工具
- Kubernetes配置
- CI/CD流水线工具
- 监控和日志工具

### 4. fullstack-lead-agent

**职责**: 全栈技术负责人
- 整体架构设计和技术选型
- 代码质量控制和审查
- 项目进度管理和协调
- 技术文档编写和维护

**负责模块**:
- 整体项目架构
- 技术选型和框架
- 代码规范和最佳实践
- 项目文档和API文档

**开发工具**:
- 架构设计工具
- 代码审查工具
- 项目管理工具
- 文档生成工具

## 协作方式

### 工作流程
1. **需求分析**: fullstack-lead-agent 负责需求分析和架构设计
2. **任务分配**: 根据需求将任务分配给相应的Agent
3. **并行开发**: 各Agent负责各自模块的开发工作
4. **代码审查**: fullstack-lead-agent 负责代码质量审查
5. **集成测试**: devops-engineer-agent 负责CI/CD和部署

### 沟通机制
- 定期技术会议：每周进行技术进度同步
- 代码审查：所有代码变更都需要经过审查
- 文档同步：及时更新技术文档和API文档
- 问题跟踪：使用Issue跟踪系统管理问题和任务

## 技术栈选择

### 前端技术栈
- **HTML/CSS/JavaScript**: 基础前端技术
- **Gin Template**: Go模板引擎
- **Bootstrap**: CSS框架
- **响应式设计**: 支持移动端

### 后端技术栈
- **Go**: 主要开发语言
- **Gin**: Web框架
- **Viper**: 配置管理
- **Logrus**: 日志管理
- **Cron**: 定时任务

### DevOps技术栈
- **Docker**: 容器化
- **GitHub Actions**: CI/CD
- **Makefile**: 构建工具
- **Git**: 版本控制