# 简聊（SimpleChat）核心 MVP 系统 Spec

## Why
依据《简聊开发手册》第一、二章需求，从零构建一个"三层分离"的实时聊天系统——客户端（Web）、服务端、数据库三者独立分层、独立目录、独立部署。以最简可用的核心 IM 功能为目标，作为课程项目的基础版本。

## What Changes
- 在 `client/`、`server/`、`database/` 三个独立目录分别建设 Web 客户端、Go 服务端、数据库脚本，三层严格分离。
- 客户端：Vue3 + Vite + Pinia + Vue Router + WebSocket（浏览器 Web 端优先，Electron 封装留待后续）。
- 服务端：Go + Gin + GORM，RESTful API（业务数据）+ WebSocket（实时消息）。
- 数据库：MySQL 初始化脚本（10 张核心表）、Redis 缓存键设计、MinIO 文件存储说明。
- **BREAKING**：全新空项目，无既有代码需要迁移。

## Impact
- Affected code：`client/`、`server/`、`database/` 三个空目录，全部新建。
- 基础设施：docker-compose 编排 MySQL + Redis + MinIO（开发环境）。

---

## ADDED Requirements

### Requirement: 系统三层分离架构
系统 SHALL 以 `client`/`server`/`database` 三个独立单元组织，三者通过 HTTP/HTTPS 与 WebSocket/WSS 通信，各自可独立启动、独立部署。

#### Scenario: 分层构建
- **WHEN** 开发者分别进入三个目录
- **THEN** 每个目录可独立安装依赖、启动服务，互不耦合

### Requirement: 服务端认证与账户（用户服务）
服务端 SHALL 提供注册、登录、登出、修改密码、资料编辑；密码 bcrypt 哈希（强度 10）存储；登录成功签发 JWT（默认 7 天，支持"记住我" 30 天）；连续 5 次失败锁定账号 15 分钟。

#### Scenario: 注册并登录
- **WHEN** 用户使用账号+密码注册并登录
- **THEN** 返回 JWT 令牌，客户端后续请求携带该令牌，并建立 WebSocket 连接

#### Scenario: 登录失败锁定
- **WHEN** 连续 5 次密码错误
- **THEN** 账号锁定 15 分钟，期间拒绝登录

### Requirement: 联系人管理（关系服务）
服务端 SHALL 支持按昵称/UUID/手机号模糊搜索用户、发起好友申请、同意/拒绝、删除好友、好友备注。

#### Scenario: 添加并备注好友
- **WHEN** 用户搜索到另一用户并发出好友申请，对方同意
- **THEN** 双方建立双向好友关系并可设置备注名

### Requirement: 单聊与已读回执（消息服务）
服务端 SHALL 通过 WebSocket 双向实时推送文本/表情/图片/文件消息；消息持久化 MySQL；支持送达/已读状态。

#### Scenario: 在线发送文本消息
- **WHEN** 发送方在线发送文本消息
- **THEN** 消息经 WebSocket 实时送达接收方，并落库持久化

#### Scenario: 已读回执
- **WHEN** 接收方打开会话并阅览消息
- **THEN** 发送方侧消息状态更新为"已读"

### Requirement: 群聊管理（消息/群组服务）
服务端 SHALL 支持创建群、解散群、群成员管理（邀请/移除/退出）、群资料（名称/公告/头像）、群内消息（同单聊）。

#### Scenario: 创建并邀请成员
- **WHEN** 用户创建群组并邀请成员
- **THEN** 群组建立，被邀请成员加入该群会话并可收发群消息

### Requirement: 会话管理（会话服务）
服务端 SHALL 提供会话列表（按最后消息时间降序）、未读数、置顶（≤10）、免打扰、删除会话（仅移出列表）。

#### Scenario: 会话排序与置顶
- **WHEN** 用户收到新消息或置顶会话
- **THEN** 会话列表按时间排序，置顶会话保持在列表最上方

### Requirement: 文件传输（文件服务）
服务端 SHALL 支持图片（jpg/png/gif/webp ≤10MB）和文件（任意格式 ≤100MB）上传，禁止可执行文件；返回文件 URL。

#### Scenario: 上传图片并发送
- **WHEN** 用户选择图片上传并作为消息发送
- **THEN** 文件存入对象存储，接收方可预览/下载

### Requirement: 消息搜索（搜索服务）
服务端 SHALL 支持在用户会话范围内按关键词搜索消息，结果按时间倒序返回，含所属会话与发送人。

#### Scenario: 关键词搜索
- **WHEN** 用户在搜索框输入关键词
- **THEN** 返回匹配消息列表，可按时间排序并定位到原会话

### Requirement: 数据库会话（数据存储层）
`database/` 目录 SHALL 提供 MySQL 初始化 SQL，含 user、friendship、group、group_member、conversation、conversation_member、message、message_status、offline_message、file_record 共 10 表及索引（InnoDB、UTF8MB4）。

#### Scenario: 初始化数据库
- **WHEN** 在空 MySQL 实例上执行初始化脚本
- **THEN** 建库建表成功，含联合唯一索引与复合索引

### Requirement: 缓存与文件存储（数据存储层）
系统 SHALL 使用 Redis 缓存会话列表、在线状态、令牌、未读计数；使用 MinIO（S3 兼容）存储图片与文件。

#### Scenario: 配置缓存与对象存储
- **WHEN** 启动开发环境
- **THEN** docker-compose 拉起 MySQL + Redis + MinIO，服务端可连接

---

## REMOVED Requirements
本阶段为全新 MVP，不实现手册中的：管理后台、消息撤回/转发、分片上传/断点续传、离线消息漫游扩展、毫秒级性能压测、Electron 桌面封装（均标记为后续任务）。