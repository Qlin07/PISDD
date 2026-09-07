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

## 核心能力（MVP）

用户认证、联系人管理、单聊 + 群聊（文本/图片/文件）、会话管理、消息搜索、系统设置。

详细需求与设计见《简聊开发手册》第一、二章。