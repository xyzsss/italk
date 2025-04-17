

# Go Gin WebSocket 聊天
这是一个基于 WebSocket 的局域网聊天室应用，使用 Go 语言和 Gin 框架开发。

## 功能特点

- 基于 WebSocket 的实时聊天功能
- 用户可自定义用户名
- 支持发送文本消息、图片（Base64编码）和表情
- 查看历史消息
- 搜索消息（按用户名、IP地址、消息内容等）
- 查看在线用户列表
- 聊天室统计信息
- 使用 SQLite 数据库存储数据

## 技术栈

- 后端：Go + Gin + Gorilla WebSocket
- 前端：HTML + CSS + JavaScript
- 数据库：SQLite

## 运行项目

### 前提条件

- 安装 Go (1.15+)
- 安装 SQLite

### 步骤

1. 克隆项目

```bash
git clone https://github.com/yourusername/go-gin-websocket-msg.git
cd go-gin-websocket-msg
```

2. 安装依赖

```bash
go mod tidy
```

3. 运行项目

```bash
go run main.go
```

4. 在浏览器中访问

```
http://localhost:8080
```

## 代码结构

```
.
├── controllers/       # 控制器
├── models/            # 数据模型
├── static/            # 静态资源
│   ├── css/           # CSS 样式
│   ├── js/            # JavaScript 脚本
│   └── images/        # 图片资源
├── templates/         # HTML 模板
├── utils/             # 工具函数
├── main.go            # 主程序入口
├── go.mod             # Go 模块文件
└── README.md          # 项目说明
```

## 开发计划

- [ ] 支持更多类型的消息
- [ ] 添加消息撤回功能
- [ ] 添加私聊功能
- [ ] 添加文件传输功能
- [ ] 添加用户身份验证

## 许可证

MIT License


