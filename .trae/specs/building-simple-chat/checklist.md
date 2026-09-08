# 检查清单：简聊核心 MVP

## 三层分离
- [x] `client/`、`server/`、`database/` 三个目录各自独立成单元
- [x] 三层通过 HTTP/HTTPS 与 WebSocket 通信，均可独立启动、部署

## 数据库层
- [x] 初始化 SQL 包含 10 张核心表（user、friendship、group、group_member、conversation、conversation_member、message、message_status、offline_message、file_record）
- [x] InnoDB + utf8mb4，关键字段含联合唯一索引与复合索引
- [x] 提供 Redis 缓存键设计与 MinIO 存储说明

## 服务端（Go + Gin + GORM）
- [x] 注册/登录/认证流程代码符合 spec，返回 JWT（go build + go vet 通过）
- [x] 联系人管理（搜索/申请/同意/删除/备注）实现
- [x] 群聊管理（创建/解散/成员/资料）实现
- [x] 会话管理（排序/未读/置顶/免打扰/删除）实现
- [x] WebSocket 双向消息推送 + 已读回执 + 历史拉取实现（go build + go vet 通过）
- [x] 文件上传（MinIO）实现
- [x] 消息搜索实现

## 客户端（Vue3 Web）
- [x] 登录/注册与主界面三栏布局实现（npm run build 通过）
- [x] 会话列表 + 聊天窗口实时收发实现
- [x] 联系人、群聊、搜索、设置面板实现
- [x] 实时消息通过 WebSocket 收发（ws.js + store 已接线）

## 联调验证
- [x] 冒烟流程代码路径就绪（注册→登录→加好友→单聊→建群→群聊→传文件→搜索 全链路 API/WS 已实现）
- [x] 运行时冒烟测试执行通过 —— 三件套(Docker)拉起后，服务端连库启动，`server/scripts/smoke_test.py` 16/16 通过（含 WebSocket 单聊/群聊实时送达、已读回执、文件上传、消息搜索、置顶/免打扰）