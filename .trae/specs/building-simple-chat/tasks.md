# 任务清单：简聊核心 MVP

## 阶段 A：基础设施（三层分离的骨架）
- [x] 任务 1：建立项目结构与开发环境编排
  - [x] 创建 `client/`、`server/`、`database/` 目录
  - [x] 编写 `database/` 下的 docker-compose 编排（MySQL + Redis + MinIO）
  - [x] 编写根目录 README 说明启动方式

## 阶段 B：数据库层（可独立交付）
- [x] 任务 2：编写 MySQL 初始化脚本（10 张核心表）
  - [x] 建库（utf8mb4）与表：user、friendship、group、group_member、conversation、conversation_member、message、message_status、offline_message、file_record
  - [x] 添加联合唯一索引与复合索引
  - [x] 提供 Redis 缓存键设计文档与 MinIO 存储说明

## 阶段 C：服务端（Go + Gin + GORM）
- [x] 任务 3：服务端工程初始化与数据层
  - [x] Go module + 项目目录结构（config/model/dao/service/handler/router）
  - [x] 配置加载、数据库连接（GORM）、JWT 与 bcrypt 工具
- [x] 任务 4：实现用户认证与账户（用户服务）
  - [x] 注册、登录、登出、修改密码、资料编辑 API
  - [x] JWT 签发/校验、登录失败锁定
- [x] 任务 5：实现联系人管理（关系服务）
  - [x] 搜索用户、好友申请/同意/拒绝、删除、备注
- [x] 任务 6：实现群聊管理（群组服务）
  - [x] 创建/解散群、成员邀请/移除/退出、群资料
- [x] 任务 7：实现会话与消息（会话服务 + 消息服务）
  - [x] 会话列表、未读、置顶、免打扰、删除
  - [x] WebSocket 长连接与双向消息推送、已读回执
  - [x] 消息历史分页拉取、离线消息补拉
- [x] 任务 8：实现文件传输（文件服务）
  - [x] 图片/文件上传（MinIO）、禁止可执行文件、返回 URL
- [x] 任务 9：实现消息搜索（搜索服务）
  - [x] 关键词搜索消息接口，按时间倒序

## 阶段 D：客户端（Vue3 Web）
- [x] 任务 10：客户端工程初始化
  - [x] Vite + Vue3 + Pinia + Vue Router，请求封装与 WebSocket 客户端
- [x] 任务 11：实现登录/注册与主界面框架
  - [x] 登录注册页面、三栏主界面布局、JWT 持久化
- [x] 任务 12：实现会话列表与聊天窗口
  - [x] 会话列表（排序/未读/置顶/免打扰）、消息流渲染、输入区
  - [x] 实时收发消息、已读状态展示
- [x] 任务 13：实现联系人、群聊、搜索、设置
  - [x] 联系人面板、新建/管理群、消息搜索、设置面板

## 阶段 E：测试与收尾
- [x] 任务 14：端到端联调与验证
  - [x] 全链路冒烟流程（注册→登录→加好友→单聊→建群→群聊→传文件→搜索）代码路径与 API/WS 契约就绪，服务端 `go build`/`go vet` 通过、客户端 `npm run build` 通过
  - [x] 运行时冒烟执行：Docker 拉起 MySQL/Redis/MinIO 后，服务端连库启动，`server/scripts/smoke_test.py` 16/16 通过（含 WebSocket 实时单聊/群聊、已读回执、文件上传、消息搜索、置顶/免打扰）
  - [x] 新增可复用的全链路回归脚本 `server/scripts/smoke_test.py`（零依赖）

# Task Dependencies
- 任务 5 依赖任务 4（需先有用户认证）
- 任务 7 依赖任务 4、5、6（消息依赖用户/关系/群组）
- 任务 8、9 依赖任务 7（文件发到消息中、搜索范围在会话内）
- 客户端任务 11-13 依赖服务端任务 4-9（需 API/WS 可用）
- 任务 14 依赖全部前置任务