# 图片上传功能实现总结

## 概述

本次更新为 go-mcp-demo 项目添加了图片上传功能，使聊天助手能够接收并分析图片内容，支持多模态 AI 模型（如 Qwen3-VL）。

## 主要改动

### 1. IDL 定义修改

**文件**: `idl/api.thrift`

在 `ChatRequest` 和 `ChatSSEHandlerRequest` 结构体中添加了可选的图片字段：

```thrift
struct ChatRequest{
    1: string message(api.body="message")
    2: optional binary image(api.form="image", api.file_name="image")
}

struct ChatSSEHandlerRequest {
    1: string message(api.query="message")
    2: optional binary image(api.form="image", api.file_name="image")
}
```

**说明**: 
- 使用 `optional binary` 类型存储图片二进制数据
- 使用 `api.form="image"` 指定表单字段名
- 使用 `api.file_name="image"` 指定文件字段名

---

### 2. 模型层自动生成

**文件**: `api/model/api/api.go`

使用 `hz update` 命令重新生成模型代码，自动添加了 `Image` 字段：

```go
type ChatRequest struct {
    Message string `thrift:"message,1" form:"message" json:"message" query:"message"`
    Image   []byte `thrift:"image,2,optional" form:"image" json:"image,omitempty" query:"image"`
}
```

**命令**: 
```bash
hz update -idl "./idl/api.thrift"
```

---

### 3. Handler 层实现文件上传

**文件**: `api/handler/api/api_service.go`

在 `Chat()` 和 `ChatSSE()` 函数中添加了文件上传处理逻辑：

```go
// Chat 函数
func Chat(ctx context.Context, c *app.RequestContext) {
    var err error
    var req api.ChatRequest
    err = c.BindAndValidate(&req)
    if err != nil {
        c.String(consts.StatusBadRequest, err.Error())
        return
    }

    // 处理图片上传
    var imageData []byte
    file, err := c.FormFile("image")
    if err == nil && file != nil {
        // 打开上传的文件
        src, err := file.Open()
        if err != nil {
            pack.RespError(c, err)
            return
        }
        defer src.Close()

        // 读取文件内容
        imageData, err = io.ReadAll(src)
        if err != nil {
            pack.RespError(c, err)
            return
        }
    }

    resp := new(api.ChatResponse)
    msg, err := host.NewHost(ctx, clientSet).Chat(1, req.Message, imageData)
    if err != nil {
        pack.RespError(c, err)
        return
    }
    resp.Response = msg
    pack.RespData(c, resp)
}
```

**说明**:
- 使用 `c.FormFile("image")` 获取上传的文件
- 使用 `io.ReadAll()` 读取文件内容到字节数组
- 将 `imageData` 传递给业务逻辑层

---

### 4. 业务逻辑层适配

#### 4.1 主聊天函数修改

**文件**: `internal/host/chat.go`

修改 `Chat()` 函数签名，添加 `imageData` 参数，并根据 AI 模式分发到不同的实现：

```go
func (h *Host) Chat(id int64, msg string, imageData []byte) (string, error) {
    // 如果是远程模式（OpenAI），使用 ChatOpenAI
    if config.AiProvider.Mode == constant.AiProviderModeRemote {
        return h.ChatOpenAI(id, msg, imageData)
    }

    // 本地模式（Ollama）的原有逻辑
    userHistory := history[id]
    if userHistory == nil {
        userHistory = []ai_provider.Message{}
    }

    // 构建用户消息
    userMsg := ai_provider.Message{Role: "user", Content: msg}
    
    // 如果有图片数据，转换为base64并添加到消息中
    if len(imageData) > 0 {
        base64Image := base64.StdEncoding.EncodeToString(imageData)
        userMsg.Images = []string{base64Image}
    }
    
    // 将当前用户消息加入历史
    userHistory = append(userHistory, userMsg)
    
    // ... 其余逻辑保持不变
}
```

#### 4.2 OpenAI 模式实现

**文件**: `internal/host/chat_openai.go`

新增 `ChatOpenAI()` 函数，支持图片上传和多模态消息：

```go
// ChatOpenAI 非流式OpenAI聊天，支持图片和工具调用
func (h *Host) ChatOpenAI(id int64, msg string, imageData []byte) (string, error) {
    // 历史（OpenAI）
    hist := historyOpenAI[id]
    if hist == nil {
        hist = []openai.ChatCompletionMessageParamUnion{}
    }
    
    // 构建用户消息
    if len(imageData) > 0 {
        // 如果有图片，使用多模态消息格式
        base64Image := base64.StdEncoding.EncodeToString(imageData)
        
        // 创建文本部分
        textPart := openai.ChatCompletionContentPartUnionParam{
            OfText: &openai.ChatCompletionContentPartTextParam{
                Type: "text",
                Text: msg,
            },
        }
        
        // 创建图片部分
        imagePart := openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
            URL: "data:image/jpeg;base64," + base64Image,
        })
        
        // 创建用户消息
        hist = append(hist, openai.ChatCompletionMessageParamUnion{
            OfUser: &openai.ChatCompletionUserMessageParam{
                Role: "user",
                Content: openai.ChatCompletionUserMessageParamContentUnion{
                    OfArrayOfContentParts: []openai.ChatCompletionContentPartUnionParam{
                        textPart,
                        imagePart,
                    },
                },
            },
        })
    } else {
        // 纯文本消息
        hist = append(hist, openai.UserMessage(msg))
    }

    // 工具（OpenAI 版）- 如果有图片则不使用工具（vision模型可能不支持）
    var tools []openai.ChatCompletionToolUnionParam
    if len(imageData) == 0 {
        tools = h.mcpCli.ConvertToolsToOpenAI()
    }

    round := 0
    for {
        round++
        if round > maxToolRounds {
            historyOpenAI[id] = hist
            return "已达到工具调用轮次上限", nil
        }

        // 调用OpenAI API
        params := openai.ChatCompletionNewParams{
            Model:    openai.ChatModel(config.AiProvider.Model),
            Messages: hist,
            Tools:    tools,
        }
        if config.AiProvider.Options.MaxTokens != nil {
            params.MaxTokens = openai.Int(int64(*config.AiProvider.Options.MaxTokens))
        }
        if config.AiProvider.Options.Temperature != nil {
            params.Temperature = openai.Float(*config.AiProvider.Options.Temperature)
        }
        if config.AiProvider.Options.TopP != nil {
            params.TopP = openai.Float(*config.AiProvider.Options.TopP)
        }
        
        resp, err := h.aiProviderCli.ChatOpenAI(h.ctx, params)
        if err != nil {
            logger.Errorf("ChatOpenAI API error: %v", err)
            return "", err
        }

        logger.Infof("ChatOpenAI response: choices=%d", len(resp.Choices))
        if len(resp.Choices) == 0 {
            logger.Errorf("ChatOpenAI: no choices in response")
            return "模型返回为空", nil
        }
        
        logger.Infof("ChatOpenAI finish_reason=%s, content=%s, tool_calls=%d", 
            resp.Choices[0].FinishReason, resp.Choices[0].Message.Content, len(resp.Choices[0].Message.ToolCalls))

        // 检查是否需要工具调用
        if resp.Choices[0].FinishReason != "tool_calls" || len(resp.Choices[0].Message.ToolCalls) == 0 {
            // 无工具调用，返回模型回复
            content := resp.Choices[0].Message.Content
            hist = append(hist, openai.AssistantMessage(content))
            historyOpenAI[id] = hist
            return content, nil
        }

        // 处理工具调用逻辑...
    }
}
```

**关键点**:
- 图片数据通过 base64 编码后使用 `data:image/jpeg;base64,` 前缀
- 使用 OpenAI 的多模态消息格式（`OfArrayOfContentParts`）
- 当有图片时不使用工具调用（某些视觉模型不支持）
- 添加了 `max_tokens`、`temperature`、`top_p` 等参数支持

#### 4.3 流式聊天支持

**文件**: `internal/host/chat_openai.go`

同样更新了 `StreamChatOpenAI()` 函数以支持图片：

```go
func (h *Host) StreamChatOpenAI(
    ctx context.Context,
    id int64,
    userMsg string,
    imageData []byte,
    emit func(event string, v any) error,
) error {
    // 历史（OpenAI）
    hist := historyOpenAI[id]
    if hist == nil {
        hist = []openai.ChatCompletionMessageParamUnion{}
    }
    
    // 构建用户消息（与非流式版本相同的图片处理逻辑）
    if len(imageData) > 0 {
        // 多模态消息格式...
    } else {
        hist = append(hist, openai.UserMessage(userMsg))
    }
    
    // ... 流式处理逻辑
}
```

---

### 5. 配置文件更新

**文件**: `config/config.yaml`

更新 AI 提供商配置以支持多模态模型：

```yaml
ai_provider:
  mode: "remote"
  base_url: "http://127.0.0.1:11434"
  model: "qwen3-vl-plus"  # 使用视觉模型
  remote:
    provider: "aliyuncs"
    base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
    api_key: "YOUR_API_KEY_HERE"  # 替换为实际的 API Key

  options:
    request_timeout: "30s"
    keep_alive: "5m"
    temperature: 0.2
    top_p: 0.9
    top_k: 40
    max_tokens: 1024
    extra: {}

mcp:
  server_name: "http.mcp.demo"
  transport: "http"
  http:
    base_url: "http://127.0.0.1:10002/mcp"
```

---

### 6. 基础设施改进

#### 6.1 超时调整

**文件**: `pkg/constant/mcp.go`

增加 MCP 客户端初始化超时时间：

```go
const (
    MCPTransportStdio          = "stdio"
    MCPTransportSSE            = "sse"
    MCPTransportHTTP           = "http"
    MCPClientInitTimeout       = 30 * time.Second  // 从 5s 增加到 30s
    MCPDefaultCallTimeout      = 30 * time.Second
    MCPServerHeartbeatInterval = 25 * time.Second
    // ...
)
```

#### 6.2 调试日志

**文件**: `pkg/base/options.go`

添加 MCP 客户端创建日志：

```go
case config.Registry.Provider == constant.RegistryProviderNone:
    if config.MCP.HTTP.BaseURL == "" {
        log.Fatalf("missing MCP HTTP BaseURL while registry provider is 'none'")
    }
    log.Printf("Creating MCP client with URL: %s", config.MCP.HTTP.BaseURL)
    mcpCli, err := mcp_client.NewMCPClient(config.MCP.HTTP.BaseURL)
    if err != nil {
        log.Fatalf("failed to create http mcp client: %s", err)
    }
    log.Printf("MCP client created successfully")
    clientSet.MCPCli = mcpCli
```

**文件**: `cmd/host/main.go`

添加启动过程日志：

```go
func init() {
    flag.Parse()
    config.Load(*configPath, serviceName)
    logger.Init(serviceName, config.GetLoggerLevel())
    logger.Infof("About to initialize API handlers")
    api.Init()
    logger.Infof("API handlers initialized successfully")
}

func main() {
    logger.Infof("main() function started")
    // ...
    logger.Infof("About to call GetAvailablePort()")
    listenAddr, err := utils.GetAvailablePort()
    // ...
    logger.Infof("Got listen address: %s", listenAddr)
    logger.Infof("Creating Hertz server")
    // ...
    logger.Infof("Server is about to start on %s", listenAddr)
    h.Spin()
    logger.Infof("Server stopped")
}
```

---

## 测试验证

### 启动服务

```bash
# 启动 MCP 本地服务
cd /root/projects/test/go-mcp-demo
export PATH=$PATH:~/go/bin
nohup go run cmd/mcp_local/main.go -cfg config/config.yaml > /tmp/mcp_local.log 2>&1 &

# 启动 Host API 服务
nohup go run cmd/host/main.go -cfg config/config.yaml > /tmp/host.log 2>&1 &
```

## 技术要点

### 1. 多模态消息格式

OpenAI API 支持多模态输入，需要使用 `content` 数组格式：

```json
{
  "role": "user",
  "content": [
    {
      "type": "text",
      "text": "请描述这张图片"
    },
    {
      "type": "image_url",
      "image_url": {
        "url": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
      }
    }
  ]
}
```

### 2. Base64 编码

- 图片数据需要 base64 编码
- 使用 data URI scheme: `data:image/jpeg;base64,{base64_string}`
- Go 标准库: `base64.StdEncoding.EncodeToString(imageData)`

### 3. 工具调用限制

某些视觉模型（如 Qwen3-VL）在处理图片时可能不支持工具调用，因此当检测到图片上传时，禁用工具调用功能。

### 4. 文件上传处理

- 使用 Hertz 的 `FormFile()` 方法获取上传文件
- 支持 `multipart/form-data` 编码
- 同时支持文本和文件字段

---

## 依赖项

- **CloudWeGo Hertz**: HTTP 框架
- **OpenAI Go SDK v2**: OpenAI API 客户端
- **Apache Thrift**: IDL 定义
- **hz**: Hertz 代码生成工具
- **thriftgo**: Thrift 编译器

---

## 注意事项

1. **API Key 安全**: 配置文件中的 API Key 应使用环境变量或密钥管理系统
2. **文件大小限制**: 需要在 Hertz 配置中设置适当的 `MaxRequestBodySize`
3. **超时设置**: 视觉模型处理时间较长，建议增加请求超时时间
4. **图片格式**: 支持常见图片格式（JPEG、PNG、GIF 等）
5. **Base64 性能**: 大图片会显著增加请求体积，建议限制图片大小

---

## 后续改进建议

1. 添加图片格式验证
2. 实现图片大小限制
3. 支持图片 URL 输入（而非仅上传）
4. 优化 base64 编码性能
5. 添加图片预处理（压缩、裁剪等）
6. 实现图片缓存机制
7. 支持批量图片上传
8. 添加更完善的错误处理和日志

---

## 相关文件清单

### 修改的文件
- `idl/api.thrift`
- `api/handler/api/api_service.go`
- `internal/host/chat.go`
- `internal/host/chat_openai.go`
- `config/config.yaml`
- `pkg/constant/mcp.go`
- `pkg/base/options.go`
- `cmd/host/main.go`

### 自动生成的文件
- `api/model/api/api.go`
- `api/router/router_gen.go`
- 其他 hz 生成文件

---

## 版本信息

- **实现日期**: 2025-11-13
- **Go 版本**: 1.x
- **Hertz 版本**: v0.10.2
- **OpenAI SDK**: v2.7.1
- **支持的模型**: OpenAI 兼容的视觉模型
