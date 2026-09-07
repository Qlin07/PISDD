# 简聊 SimpleChat · 数据存储层

`client/`、`server/`、`database/` 三层严格分离，本目录为**数据存储层**。

## 快速启动

```bash
cd database
docker-compose up -d
```

会拉起三个服务：

| 服务 | 端口 | 账号 | 用途 |
|------|------|------|------|
| MySQL | 3306 | root / root123456（库：simplechat） | 主数据库 |
| Redis | 6379 | 无密码 | 缓存 / 会话状态 |
| MinIO | 9000(API) / 9001(控制台) | minioadmin / minioadmin123 | 文件对象存储 |

> MySQL 首次启动会自动执行 `mysql/init/01_schema.sql` 建库建表（10 张核心表）。

## 初始化 SQL

- `mysql/init/01_schema.sql`：建库 + 10 张表层结构 + 联合唯一索引 + 复合索引。

## Redis 缓存键设计

| 缓存 Key | 数据结构 | 过期时间 | 说明 |
|----------|---------|----------|------|
| `user:{user_id}` | Hash | 1 小时 | 用户基本信息缓存 |
| `session:{user_id}` | List | 30 分钟 | 用户在线会话（WebSocket 连接） |
| `token:{user_id}` | String | 7 天 | JWT 令牌 |
| `conversation:list:{user_id}` | ZSET | 10 分钟 | 用户会话列表（按时间排序） |
| `unread:{user_id}` | Hash | 实时 | 各会话未读计数 |
| `online:{user_id}` | String | 5 分钟 | 在线状态（心跳续期） |

## MinIO（对象存储）

- Bucket：`simplechat-files`
- 用途：存储图片与文件。上传文件返回 `file_url`，数据库中 `file_record` 表登记元数据。
- 客户端通过 `file_url` 直接预览/下载；服务端配置访问密钥上传。

## 目录规划

```
database/
├── docker-compose.yml        # MySQL + Redis + MinIO 编排
├── mysql/init/01_schema.sql  # 建库建表脚本
└── README.md                 # 本说明
```