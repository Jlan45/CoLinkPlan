# 产品需求文档 (PRD) - Co-Link Plan 客户端 (Gateway)

## 1. 产品概述
**Co-Link Plan Client** 是一个部署在拥有 AI 资源用户本地的轻量级守护进程。它负责主动连接服务端并保持 WebSocket 长连接，接收服务端的转发指令，将指令转换为对应大模型厂商（OpenAI 或 Claude）的真实 API 请求，并将接收到的 SSE (Server-Sent Events) 数据流实时推回给服务端。

## 2. 核心技术栈建议
* **开发语言:** Go 或 Python (Python 适合快速迭代与引入现有 AI SDK，Go 适合打包为单文件跨平台分发且内存占用低)。
* **配置文件:** YAML 格式。

## 3. 核心功能需求

### 3.1 本地配置管理
* **映射规则:** 用户需在 `config.yaml` 中配置本地拥有的 API 资源与服务端标准模型名的映射。
  ```yaml
  client_token: "user_jlan_001_secret"
  server_url: "wss://[api.yourdomain.com/ws](https://api.yourdomain.com/ws)"
  max_parallel: 3 # 本地最高并发数
  providers:
    - type: "openai"
      api_key: "sk-real-openai-key"
      base_url: "[https://api.openai.com/v1](https://api.openai.com/v1)" # 支持本地代理或其他第三方中转
      models:
        - local: "gpt-4-0125-preview"
          server_mapping: "pro-model"
    - type: "claude"
      api_key: "sk-ant-api03-xxx"
      models:
        - local: "claude-3-opus-20240229"
          server_mapping: "ultra-model"
    ```
### 3.2 协议转换与代理转发

* **OpenAI 适配器:** 接收服务端的 JSON Payload，直接透传或组装为标准 OpenAI HTTP 请求，发起长连接监听响应流。
* **Claude 适配器:** 负责将 OpenAI 格式的 Messages (如 `system`, `user`, `assistant`) 转换为 Anthropic 官方要求的格式发起请求。
* **流式封包:** 将监听到的 Chunk 数据包，封装入 WS 的 `STREAM` 消息体中，立刻上报服务端，保证调用端体验到的打字机效果无延迟。

### 3.3 健壮性与自我保护

* **断线重连:** 若服务端宕机或网络波动导致 WS 断开，客户端需启动指数退避 (Exponential Backoff) 机制进行重连 (如 2s, 4s, 8s, 16s...)。
* **本地并发控制:** 严格遵守配置中的 `max_parallel` 上限。如本地正在处理的请求数已达上限，即使服务端错误下发了任务，客户端也必须拒绝并返回 `BUSY` 状态，保护本地 API Key 不被官方封禁。
* **脱敏处理:** 客户端在控制台输出的日志中，默认不打印 Prompt 详情和完整的生成文本，仅打印 Request ID、耗时和状态码。
