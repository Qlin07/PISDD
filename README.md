# 简聊 SimpleChat

实时聊天系统课程项目，采用**三层分离**架构：客户端、服务端、数据库相互独立，各自部署。

## 架构

```
client/     表现层   Vue3 + Vite + Pinia + Vue Router + WebSocket（浏览器 Web 端）
server/     业务逻辑层  Go + Gin + GORM（RESTful API + WebSocket 消息推送）
database/   数据存储层  MySQL + Redis + MinIO（Docker 编排）
```

## 快速启动

### 1. 启动数据存储层

```bash
cd database
docker-compose up -d
```

### 2. 启动服务端

```bash
cd server
go mod tidy
go run main.go
```

### 3. 启动客户端

```bash
cd client
npm install
npm run dev
```

浏览器访问客户端地址，注册/登录后即可开始使用。

> **环境要求**：Go 1.22+、Node.js、Docker（运行数据层）。
> 首次启动 `go mod tidy`、`npm install` 安装依赖。

## 核心能力（MVP）

用户认证、联系人管理、单聊 + 群聊（文本/图片/文件）、会话管理、消息搜索、系统设置。

详细需求与设计见《简聊开发手册》第一、二章。

## 当前进度

| 模块 | 状态 |
| --- | --- |
| 三层架构（client / server / database） | ✅ 已实现并分离 |
| 数据层编排（MySQL + Redis + MinIO，Docker Compose） | ✅ 一键部署 |
| 服务端（Go + Gin + GORM） | ✅ RESTful API + WebSocket 实时推送 |
| 客户端（Vue3 + Vite + Pinia） | ✅ 登录/注册、主界面、三栏布局、聊天窗口、设置面板 |
| 认证 / 联系人 / 会话 / 消息 / 文件 / 搜索 | ✅ 已实现 |
| 注册/登录/加好友/单聊/建群/群聊/文件/搜索 | ✅ 全链路运行时冒烟测试通过 |
| 前端实时消息（发送回显 + 在线接收） | ✅ 统一消息源，即时显示、去重 |
| 前端 Web 规范合规 | ✅ 语义化 button、表单 label、aria-label、焦点态、图片 alt |
| 视觉设计（夜航暗色主题） | ✅ 银夜色蓝黑基底 + 夜航青蓝发光强调，全组件暗色适配 |

## 测试

数据层 + 服务端启动后，运行全链路回归脚本（零第三方依赖）：

```bash
python server/scripts/smoke_test.py
```

覆盖：注册 → 登录 → 加好友 → 单聊 → 建群 → 群聊 → 传文件 → 搜索 → WebSocket 实时收发。16/16 用例通过即链路正常。