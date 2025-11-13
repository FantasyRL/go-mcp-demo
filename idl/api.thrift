namespace go api
include "model.thrift"
include "openapi.thrift"
struct ChatRequest{
    1: string message(api.body="message", openapi.property='{
        title: "用户消息",
        description: "用户发送的消息内容",
        type: "string"
    }')
    2: optional binary image(api.form="image", api.file_name="image", openapi.property='{
        title: "图片文件",
        description: "可选的图片文件，支持上传图片给AI分析",
        type: "string",
        format: "binary"
    }')
}(
    openapi.schema='{
        title: "聊天请求",
        description: "包含用户消息的聊天请求",
        required: ["message"]
    }'
)

struct ChatResponse{
    1: string response(api.body="response", openapi.property='{
        title: "AI回复",
        description: "AI生成的回复内容",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "聊天响应",
        description: "包含AI回复的聊天响应",
        required: ["response"]
    }'
)

struct ChatSSEHandlerRequest{
    1: string message(api.query="message",openapi.property='{
        title: "用户消息",
        description: "用户发送的消息内容",
        type: "string"
    }')
    2: optional binary image(api.form="image", api.file_name="image", openapi.property='{
        title: "图片文件",
        description: "可选的图片文件，支持上传图片给AI分析",
        type: "string",
        format: "binary"
    }')
}(
     openapi.schema='{
         title: "流式聊天请求",
         description: "包含用户消息的流式聊天请求",
         required: ["message"]
     }'
)

struct ChatSSEHandlerResponse{
    1: string response(api.body="response", openapi.property='{
        title: "AI回复片段",
        description: "AI生成的回复片段",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "流式聊天响应",
        description: "包含AI回复片段的流式聊天响应",
        required: ["response"]
    }'
)

service ApiService {
    // 非流式对话
    ChatResponse Chat(1: ChatRequest req)(api.post="/api/v1/chat")
    // 流式对话
    ChatSSEHandlerResponse ChatSSE(1: ChatSSEHandlerRequest req)(api.get="/api/v1/chat/sse")
}