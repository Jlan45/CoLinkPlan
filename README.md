# Co-Link Plan

**Co-Link** 是一个分布式 AI 算力代理平台，让你将多台本地共享设备聚合成统一的、OpenAI 兼容的 API 端点。
```
本地共享设备 ─┐
本地共享设备 ─┼──[WebSocket]──▶ Co-Link Server ──▶ API 调用者
本地共享设备 ─┘
```

---

## 功能特性

- **OpenAI 全兼容** — 支持 Chat Completions（流式 & 非流式）、Models API
- **分布式调度** — 按并发负载动态路由，自动 failover 重试（最多 3 次）
- **零信任鉴权** — JWT 用户认证 + bcrypt 密码哈希 + API Token / Client Token 双令牌体系
- **速率限制** — 基于 Redis 的 RPM（每分钟请求数）限流
- **多 Provider 支持** — 客户端可同时接入 OpenAI 兼容接口（包括 Claude via 适配器）
- **内嵌前端** — React 前端编译后通过 `go:embed` 打包进服务端二进制，零额外依赖
- **节点惩罚机制** — 出错节点自动封禁 60 秒，避免流量持续路由到故障节点

---

## 架构overview

```
┌─────────────────────────────────────────────────────┐
│                   Co-Link Server                    │
│                                                     │
│  HTTP API (/v1/*)    Hub (调度中心)   WebSocket      │
│  ┌─────────────┐    ┌────────────┐   ┌──────────┐  │
│  │ /chat/      │───▶│ RouteCall  │──▶│ Client   │  │
│  │ completions │    │ Select     │   │ Registry │  │
│  │ /models     │    │ Failover   │   └──────────┘  │
│  └─────────────┘    └────────────┘                  │
│                                                     │
│  Auth / Rate Limit  PostgreSQL    Redis             │
└─────────────────────────────────────────────────────┘
         ▲                              ▼
   API 调用者                    本地 client 节点
   (SDK / curl)                 (GPU 机器)
```

---

## 目录结构

```
.
├── cmd/
│   ├── server/         # 服务端入口
│   └── client/         # 客户端守护进程入口
├── internal/
│   ├── server/
│   │   ├── gateway.go  # HTTP 路由处理（Chat、Models）
│   │   ├── hub.go      # WebSocket 连接调度中心
│   │   ├── client_conn.go  # 单个 client 节点连接维护
│   │   └── auth.go     # 用户注册/登录/JWT 验证
│   ├── client/
│   │   └── manager.go  # 客户端 WebSocket 连接管理和任务处理
│   ├── adapter/
│   │   ├── openai.go   # OpenAI 兼容 provider 适配器
│   │   └── claude.go   # Claude provider 适配器
│   ├── config/
│   │   ├── server_config.go   # 服务端环境变量配置
│   │   └── client_config.go   # 客户端 YAML 配置
│   ├── protocol/
│   │   ├── protocol.go # WebSocket 消息类型定义
│   │   └── models.go   # OpenAI API 请求/响应结构体
│   ├── db/             # PostgreSQL 数据访问层
│   └── limiter/        # Redis 速率限制
├── web/                # React 前端（Vite + TypeScript）
│   ├── src/
│   │   ├── pages/      # Home, Dashboard, Nodes, Login, Register
│   │   └── components/ # 共享 Navbar 组件
│   └── embed.go        # go:embed 将 dist/ 打包进二进制
├── Makefile
└── client.yaml         # 客户端配置示例
```

---

## 快速开始

### 服务端部署

#### 1. 环境依赖

- Go 1.21+
- Node.js 18+ / npm
- PostgreSQL
- Redis

#### 2. 配置环境变量

```bash
export PORT=8080
export DATABASE_URL="postgres://user:password@localhost:5432/colink?sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"
export JWT_SECRET="your-strong-secret-key"
```

#### 3. 一键编译（含前端）

```bash
make build
```

等同于依次执行：

```bash
# 编译 React 前端
cd web && npm install && npm run build

# 编译 Go 服务端（含嵌入前端）
go build -o bin/server ./cmd/server

# 编译 Go 客户端
go build -o bin/client ./cmd/client
```

#### 4. 启动服务端

```bash
./bin/server
# 默认监听 :8080
```

### 使用 Docker 部署 (推荐)

项目提供了 `docker-compose.yml`，可以一键启动全栈服务（含数据库、缓存和网关）：

```bash
docker-compose up -d --build
```

该模式下：
- **网关**：监听宿主机 `8080` 端口。
- **数据库 & Redis**：仅在 Docker 内部网络开放，不对外暴露端口，确保安全。

---

访问 `http://localhost:8080` 即可看到 Web 控制台。

---

### 客户端接入（贡献算力节点）

#### 1. 注册账号并获取 Client Token

访问 Web 控制台 → 注册账号 → 个人面板 → 复制 **Gateway Client Token**

#### 2. 创建配置文件

```yaml
# config.yaml
client_token: "client-your-token-here"
server_url: "ws://your-server:8080/ws"   # 生产环境使用 wss://
max_parallel: 3                           # 最大并发任务数

providers:
  - type: "openai"
    api_key: "sk-your-api-key"
    base_url: "https://api.openai.com/v1"  # 可选，默认 OpenAI
    models:
      - local: "gpt-4-turbo"       # 提供商侧模型名
        server_mapping: "pro-model" # 在网关中暴露的名称
```

**Provider 类型**：

| `type` | 说明 |
|--------|------|
| `openai` | OpenAI 兼容接口（含各类国内兼容服务） |
| `claude` | Anthropic Claude（通过适配器转换） |

#### 3. 启动客户端

```bash
# 从当前目录加载 config.yaml（默认）
./bin/client

# 指定配置文件路径
./bin/client -c /path/to/my-config.yaml
./bin/client --config /path/to/my-config.yaml
```

---

### API 接入（调用 AI 服务）

注册账号 → 个人面板 → 复制 **Server API Token**

```python
# Python OpenAI SDK
from openai import OpenAI

client = OpenAI(
    api_key="sk-colink-your-api-token",
    base_url="http://your-server:8080/v1",
)

# 流式调用
stream = client.chat.completions.create(
    model="pro-model",
    messages=[{"role": "user", "content": "Hello!"}],
    stream=True,
)
for chunk in stream:
    print(chunk.choices[0].delta.content or "", end="")

# 非流式调用
resp = client.chat.completions.create(
    model="pro-model",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(resp.choices[0].message.content)
```

```bash
# curl
curl -X POST http://your-server:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-colink-your-api-token" \
  -H "Content-Type: application/json" \
  -d '{"model":"pro-model","stream":true,"messages":[{"role":"user","content":"Hello!"}]}'
```

---

## API 参考

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/v1/chat/completions` | POST | API Token | Chat 对话（流式 & 非流式） |
| `/v1/models` | GET | API Token | 列出当前在线的所有模型 |
| `/v1/models/:model` | GET | API Token | 查询单个模型信息 |
| `/ws` | WebSocket | Client-Token Header | 客户端节点接入 |
| `/api/auth/register` | POST | — | 注册账号 |
| `/api/auth/login` | POST | — | 登录获取 JWT |
| `/api/user/me` | GET | JWT | 获取当前用户信息和 Tokens |
| `/api/nodes` | GET | — | 获取活跃节点列表（公开） |

---

## WebSocket 协议

服务端与客户端节点通过 WebSocket 通信，消息格式：

```json
{ "type": "MESSAGE_TYPE", "data": { ... } }
```

| Type | 方向 | 说明 |
|------|------|------|
| `REGISTER` | Client → Server | 注册节点，上报支持的模型和最大并发数 |
| `CALL` | Server → Client | 分配推理任务 |
| `STREAM` | Client → Server | 返回流式 chunk |
| `FINISH` | Client → Server | 任务完成 |
| `ERROR` | 双向 | 任务级或连接级错误 |

---

## Makefile 命令

```bash
make build        # 编译前端 + 服务端 + 客户端
make build-web    # 仅编译 React 前端
make build-server # 仅编译服务端 Go 二进制
make build-client # 仅编译客户端 Go 二进制
make clean        # 清理 bin/ 和 web/dist/
```

---

## License

MIT
