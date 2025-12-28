# NOFX 运维文档

> 本文档为 NOFX 项目的运维指南，帮助新程序员快速了解系统架构、部署流程和日常维护操作。

---

## 目录

1. [系统概述](#系统概述)
2. [技术架构](#技术架构)
3. [环境准备](#环境准备)
4. [部署指南](#部署指南)
5. [日常运维](#日常运维)
6. [监控与日志](#监控与日志)
7. [故障排查](#故障排查)
8. [备份与恢复](#备份与恢复)
9. [安全配置](#安全配置)
10. [常见问题](#常见问题)

---

## 系统概述

### 项目简介

NOFX 是一个 AI 驱动的加密货币自动交易系统，支持多个 AI 模型和交易所。

### 核心功能

- **多 AI 模型支持**: DeepSeek, Qwen, OpenAI, Claude, Gemini, Grok, Kimi, API Gateway
- **多交易所支持**: Binance, Bybit, OKX, Bitget, Hyperliquid, Aster DEX, Lighter
- **策略工作室**: 可视化策略构建器
- **AI 辩论场**: 多 AI 模型辩论交易决策
- **竞争模式**: 多 AI 交易员实时竞争
- **回测实验室**: 策略回测与验证

### 访问地址

- **前端界面**: http://72.61.125.104:3001
- **后端 API**: http://72.61.125.104:8080
- **健康检查**: http://72.61.125.104:8080/api/health

---

## 技术架构

### 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 后端 | Go | 1.21+ |
| 前端 | React | 18+ |
| 前端语言 | TypeScript | 5.0+ |
| 构建工具 | Vite | 6.x |
| 数据库 | SQLite | 嵌入式 |
| 容器 | Docker | - |
| 容器编排 | Docker Compose | - |

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                         用户浏览器                            │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  Frontend Container (nginx + React Static Files)            │
│  - Port: 3000 → 80                                           │
│  - 反向代理后端 API                                           │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  Backend Container (Go API Server)                          │
│  - Port: 8080                                                │
│  - SQLite Database: /app/data/data.db                        │
│  - AI 模型客户端                                              │
│  - 交易所客户端                                              │
└─────────────────────────────────────────────────────────────┘
```

### 目录结构

```
nofx/
├── api/              # API 路由和处理
├── auth/             # 认证模块
├── backtest/         # 回测系统
├── config/           # 配置管理
├── crypto/           # 加密模块
├── data/             # 数据目录 (SQLite 数据库)
├── debate/           # AI 辩论系统
├── docker/           # Docker 构建文件
│   ├── Dockerfile.backend
│   └── Dockerfile.frontend
├── hook/             # 交易钩子
├── logger/           # 日志模块
├── mcp/              # AI 模型客户端
├── store/            # 数据存储层
├── trader/           # 交易员实现
├── web/              # 前端代码
│   ├── src/
│   │   ├── components/  # React 组件
│   │   ├── pages/       # 页面组件
│   │   ├── stores/      # 状态管理
│   │   ├── lib/         # 工具库
│   │   └── types.ts     # TypeScript 类型定义
│   └── package.json
├── .env               # 环境变量配置
├── docker-compose.yml # Docker Compose 配置
├── Makefile          # 开发命令
└── main.go           # 后端入口
```

---

## 环境准备

### 系统要求

- **操作系统**: Linux (推荐 Ubuntu 20.04+ / CentOS 8+)
- **内存**: 最低 2GB，推荐 4GB+
- **磁盘**: 最低 10GB 可用空间
- **网络**: 需要访问外部 API (AI 服务商、交易所)

### 软件依赖

```bash
# Docker
curl -fsSL https://get.docker.com | sh

# Docker Compose
docker compose version

# Git
apt install git
```

---

## 部署指南

### 1. 克隆代码

```bash
cd /root/nofx  # 或你的项目目录
git pull origin dev  # 或你的分支
```

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件
nano .env
```

### 必填配置项

```bash
# JWT 密钥 (生成: openssl rand -base64 32)
JWT_SECRET=your-jwt-secret-here

# 数据加密密钥 (生成: openssl rand -base64 32)
DATA_ENCRYPTION_KEY=your-encryption-key-here

# RSA 私钥 (生成: openssl genrsa 2048)
RSA_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----
... 你的密钥 ...
-----END RSA PRIVATE KEY-----

# 传输加密 (HTTP 环境设为 false)
TRANSPORT_ENCRYPTION=false
```

### 3. 构建并启动服务

```bash
# 构建镜像
docker compose build

# 启动服务
docker compose up -d

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f
```

### 4. 验证部署

```bash
# 检查后端健康状态
curl http://localhost:8080/api/health

# 检查支持的模型
curl http://localhost:8080/api/supported-models

# 访问前端
# 浏览器打开: http://your-server-ip:3001
```

---

## 日常运维

### 服务管理命令

```bash
# 启动服务
docker compose up -d

# 停止服务
docker compose down

# 重启服务
docker compose restart

# 重启后端
docker compose restart nofx

# 重启前端
docker compose restart nofx-frontend

# 查看日志
docker compose logs -f nofx          # 后端日志
docker compose logs -f nofx-frontend # 前端日志
```

### 代码更新流程

```bash
# 1. 拉取最新代码
git pull origin dev

# 2. 重新构建 (如代码有变化)
docker compose build --no-cache nofx          # 后端
docker compose build --no-cache nofx-frontend # 前端

# 3. 重启服务
docker compose up -d
```

### 数据库操作

```bash
# 进入后端容器
docker exec -it nofx-trading sh

# 数据库位置
cd /app/data

# 备份数据库
docker cp nofx-trading:/app/data/data.db ./backup_$(date +%Y%m%d_%H%M%S).db

# 恢复数据库
docker cp ./backup_xxx.db nofx-trading:/app/data/data.db
docker compose restart nofx
```

---

## 监控与日志

### 日志位置

- **后端日志**: `docker compose logs nofx`
- **前端日志**: `docker compose logs nofx-frontend`
- **数据库**: `/app/data/data.db` (容器内)

### 关键监控指标

```bash
# 容器状态
docker ps

# 容器资源使用
docker stats nofx-trading nofx-frontend

# 磁盘使用
du -sh /root/nofx/data/

# 数据库大小
docker exec nofx-trading ls -lh /app/data/
```

### 健康检查端点

| 端点 | 说明 |
|------|------|
| `/api/health` | 后端健康状态 |
| `/api/config` | 系统配置状态 |

---

## 故障排查

### 后端无法启动

```bash
# 1. 检查日志
docker compose logs nofx

# 2. 检查环境变量
cat .env

# 3. 检查端口占用
netstat -tlnp | grep 8080

# 4. 检查数据库权限
docker exec nofx-trading ls -la /app/data/
```

### 前端页面无法访问

```bash
# 1. 检查容器状态
docker ps | grep nofx-frontend

# 2. 检查前端日志
docker compose logs nofx-frontend

# 3. 检查 nginx 配置
docker exec nofx-frontend cat /etc/nginx/nginx.conf
```

### AI 模型调用失败

```bash
# 1. 检查 API Key 配置
# 通过前端界面检查 AI 模型配置

# 2. 测试网络连接
docker exec nofx-trading wget -O- https://api.deepseek.com

# 3. 检查后端日志
docker compose logs nofx | grep -i "api"
```

### 交易所连接失败

```bash
# 1. 检查 API Key 配置
# 通过前端界面检查交易所配置

# 2. 检查网络连接
docker exec nofx-trading wget -O- https://fapi.binance.com

# 3. 查看交易员日志
docker compose logs nofx | grep -i "trader"
```

---

## 备份与恢复

### 自动备份脚本

创建备份脚本 `/root/nofx/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/root/nofx/backups"
mkdir -p $BACKUP_DIR

# 备份数据库
docker cp nofx-trading:/app/data/data.db $BACKUP_DIR/data_$(date +%Y%m%d_%H%M%S).db

# 保留最近 7 天的备份
find $BACKUP_DIR -name "data_*.db" -mtime +7 -delete

echo "Backup completed: $(date)"
```

### 定时备份

```bash
# 添加到 crontab
crontab -e

# 每天凌晨 3 点执行备份
0 3 * * * /root/nofx/backup.sh >> /root/nofx/backup.log 2>&1
```

### 恢复流程

```bash
# 1. 停止服务
docker compose down

# 2. 恢复数据库
docker cp ./backup_xxx.db nofx-trading:/app/data/data.db

# 3. 启动服务
docker compose up -d
```

---

## 安全配置

### 密钥生成

```bash
# JWT 密钥
openssl rand -base64 32

# 数据加密密钥
openssl rand -base64 32

# RSA 私钥
openssl genrsa 2048
```

### 防火墙配置

```bash
# 仅允许必要端口
ufw allow 80/tcp   # HTTP
ufw allow 443/tcp  # HTTPS
ufw allow 22/tcp   # SSH
ufw enable
```

### SSL/TLS 配置 (可选)

如需启用 HTTPS，建议使用 Nginx 反向代理:

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 常见问题

### Q1: 忘记管理员密码?

```bash
# 进入容器
docker exec -it nofx-trading sh

# 使用 sqlite3 操作数据库
sqlite3 /app/data/data.db

# 重置密码 (需要实现管理命令)
```

### Q2: 如何查看运行中的交易员?

通过前端界面访问 **AI 交易员** 页面，或调用 API:

```bash
curl http://localhost:8080/api/traders
```

### Q3: 如何增加新的 AI 模型?

1. 在 `mcp/` 目录下创建新的客户端
2. 在 `api/server.go` 的 `handleGetSupportedModels` 中添加模型定义
3. 重新构建并部署

### Q4: 如何添加新的交易所?

1. 在 `trader/` 目录下创建新的交易所实现
2. 实现 `Trader` 接口
3. 在 `api/server.go` 中添加交易所配置

### Q5: 数据库文件太大怎么办?

```bash
# SQLite 数据库清理
docker exec -it nofx-trading sqlite3 /app/data/data.db "VACUUM;"

# 检查数据库大小
docker exec nofx-trading ls -lh /app/data/
```

---

## 联系方式

- **项目地址**: https://github.com/your-org/nofx
- **核心团队**: Tinkle (@Web3Tinkle)
- **官方 Twitter**: [@nofx_official](https://x.com/nofx_official)
- **开发者社区**: [NOFX Developer Community](https://t.me/nofx_dev_community)

---

*最后更新: 2025-12-26*
