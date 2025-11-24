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
        title:"图片文件",
        description:"可选的图片文件，支持上传图片给AI分析",
        type:"string",
        format:"binary"
    }')
    3: string conversation_id(api.body="conversation_id", openapi.property='{
        title:"对话ID",
        description:"前端生成的UUID，多轮会话唯一标识",
        type:"string"
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
        title:"AI回复",
        description:"AI生成的回复内容",
        type:"string"
    }')
    2: optional string conversation_id(api.body="conversation_id", openapi.property='{
        title:"对话ID",
        description:"回显本轮所属的对话UUID",
        type:"string"
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
        title:"图片文件",
        description:"可选的图片文件，支持上传图片给AI分析",
        type:"file"
    }')
    3: string conversation_id(api.query="conversation_id", openapi.property='{
        title:"对话ID",
        description:"前端生成的UUID，多轮会话标识",
        type:"string"
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
        title:"AI回复片段",
        description:"AI生成的回复片段",
        type:"string"
    }')
    2: optional string conversation_id(api.body="conversation_id", openapi.property='{
        title:"对话ID UUID",
        type:"string"
    }')
}(
    openapi.schema='{
        title:"流式聊天响应",
        description:"包含AI回复片段的流式聊天响应",
        required:["response"]
    }'
)

struct GetConversationHistoryRequest {
    1: string conversation_id(api.query="conversation_id", openapi.property='{
        title:"对话ID",
        description:"要获取的对话UUID",
        type:"string"
    }')
}(
    openapi.schema='{
        title:"获取历史请求",
        description:"按UUID获取完整对话历史",
        required:["conversation_id"]
    }'
)

struct GetConversationHistoryResponse {
    1: string conversation_id(api.body="conversation_id", openapi.property='{
        title:"对话ID",
        type:"string"
    }')
    2: string messages(api.body="messages", openapi.property='{
        title:"json消息",
        type:"strig"
    }')
}(
    openapi.schema='{
        title:"获取历史响应",
        description:"返回对话的全部消息"
        required:["conversation_id","messages"]
    }'
)

struct TemplateRequest{
    1: string templateId(api.body="templateId", openapi.property='{
        title: "示范用param",
        description: "示范用param",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "示例请求",
        description: "示例请求",
        required: ["templateId"]
    }'
)

struct TemplateResponse{
    1: model.User user(api.body="user", openapi.property='{
        title: "示范用返回值",
        description: "示范用返回值",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "示例响应",
        description: "示例响应",
        required: ["user"]
    }'
)

struct SummarizeConversationRequest{
    1: string conversation_id(api.body="conversation_id", openapi.property='{
        title: "会话ID",
        description: "需要总结的会话ID",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "总结会话请求",
        description: "请求总结指定会话的内容",
        required: ["conversation_id"]
    }'
)

struct SummarizeConversationResponse{
    1: string summary(api.body="summary", openapi.property='{
        title: "会话总结",
        description: "会话内容的总结",
        type: "string"
    }')
    2: list<string> tags(api.body="tags", openapi.property='{
        title: "标签列表",
        description: "会话相关的标签",
        type: "array"
    }')
    3: string tool_calls_json(api.body="tool_calls_json", openapi.property='{
        title: "工具调用JSON",
        description: "工具调用的JSON字符串",
        type: "string"
    }')
    4: map<string,string> notes(api.body="notes", openapi.property='{
        title: "笔记",
        description: "AI 或用户针对总结写的任意键值笔记，包含文件路径等信息",
        type: "object"
    }')
}(
    openapi.schema='{
        title: "总结会话响应",
        description: "包含会话总结、标签、工具调用和笔记的响应",
        required: ["summary", "tags", "tool_calls_json", "notes"]
    }'
)

struct GetLoginDataRequest{
    1: string stu_id(api.body="stu_id", openapi.property='{
        title: "学号",
        description: "福州大学学号",
        type: "string"
    }')
    2: string password(api.body="password", openapi.property='{
        title: "密码",
        description: "用户的登录密码",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "登录请求",
        description: "包含学号和密码的登录请求",
        required: ["stu_id", "password"]
    }'
)

struct GetLoginDataResponse{
    1: string identifier(api.body="identifier", openapi.property='{
        title: "用户ID",
        description: "登录成功后返回的用户唯一标识符",
        type: "string"
    }')
    2: string cookie(api.body="cookie", openapi.property='{
        title: "会话Cookie",
        description: "登录成功后返回的会话Cookie",
        type: "string"
    }')
    3: string access_token(api.body="access_token", openapi.property='{
        title: "访问令牌",
        description: "登录成功后返回的访问令牌",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "登录响应",
        description: "包含用户ID和会话Cookie的登录响应",
        required: ["identifier", "cookie","access_token"]
    }'
)

struct GetUserInfoRequest {
}(
    openapi.schema='{
        title: "用户信息请求",
        description: "请求用户的基本信息"
    }'
)

struct GetUserInfoResponse {
    1: string user_id(api.body="user_id", openapi.property='{
        title: "用户ID",
        description: "用户的唯一标识符",
        type: "string"
    }')
    2: string username(api.body="username", openapi.property='{
        title: "用户名",
        description: "用户的登录名",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "用户信息响应",
        description: "包含用户ID和用户名的响应",
        required: ["user_id", "username"]
    }'
)

service ApiService {
    // 非流式对话
    ChatResponse Chat(1: ChatRequest req)(api.post="/api/v1/chat")
    // 流式对话
    ChatSSEHandlerResponse ChatSSE(1: ChatSSEHandlerRequest req)(api.post="/api/v1/chat/sse")
    // 示例接口 idl写好后运行make hertz-gen-api生成脚手架
    TemplateResponse Template(1: TemplateRequest req)(api.post="/api/v1/template")
    // 获取会话历史
    GetConversationHistoryResponse GetConversationHistory(1: GetConversationHistoryRequest req)(api.get="/api/v1/conversation/history")
    // 会话总结
    SummarizeConversationResponse SummarizeConversation(1: SummarizeConversationRequest req)(api.post="/api/v1/conversation/summarize")
    // 获取jwch登录数据
    GetLoginDataResponse GetLoginData(1: GetLoginDataRequest req)(api.post="/api/v1/user/login")
    // 获取用户信息
    GetUserInfoResponse GetUserInfo(1: GetUserInfoRequest req)(api.get="/api/v1/user/info")
}