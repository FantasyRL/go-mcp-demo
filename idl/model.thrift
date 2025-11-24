namespace go model
include "openapi.thrift"
struct BaseResp {
    1: i64 code (api.body="code", openapi.property='{
        title: "状态码",
        description: "响应状态码",
        type: "integer"
    }')
    2: string msg (api.body="msg", openapi.property='{
        title: "消息",
        description: "响应消息",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "基础响应",
        description: "所有响应的基础结构",
        required: ["code", "msg"]
    }'
)

struct User {
    1: string id (api.body="id", openapi.property='{
        title: "用户ID",
        description: "唯一标识用户的ID",
        type: "string"
    }')
    2: string name (api.body="name", openapi.property='{
        title: "用户名",
        description: "用户的显示名称",
        type: "string"
    }')
}(
    openapi.schema='{
        title: "用户信息",
        description: "包含用户基本信息的结构",
        required: ["id", "name"]
    }'
)

struct ConversationHistoryMessage {
    1: string role(api.body="role", openapi.property='{
        title:"角色",
        description:"user/assistant/tool",
        type:"string"
    }')
    2: string content(api.body="content", openapi.property='{
        title:"内容",
        description:"消息文本或工具结果",
        type:"string"
    }')
    3: optional string tool_name(api.body="tool_name", openapi.property='{
        title:"工具名",
        type:"string"
    }')
    4: optional list<string> images(api.body="images", openapi.property='{
        title:"图片Base64列表",
        type:"array"
    }')
}(
    openapi.schema='{
        title:"单条对话消息",
        description:"对话中的一条消息"
    }'
)