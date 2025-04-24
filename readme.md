# Go-Gin WebSocket 聊天室

一个基于Go、Gin和WebSocket的实时聊天应用，支持文本消息、图片、表情、文件发送和消息撤回功能。

## 功能特点

- 实时消息传递（基于WebSocket）
- 支持文本消息、图片、表情符号
- 支持文件上传和下载
- 消息撤回功能（8小时内）
- 用户在线状态显示
- 消息历史记录和搜索
- 跨平台支持（Windows、macOS、Linux）

## 系统要求

### 开发环境
- Go 1.19+
- 用于文件数据库模式的C编译器（CGO_ENABLED=1）

### 运行环境
应用有两种运行模式：
1. **文件数据库模式**（需CGO）：数据保存在本地SQLite文件中
2. **内存数据库模式**：无依赖，但应用关闭后数据会丢失

## 打包和运行

### 从源码构建

1. 确保您已安装Go 1.19+和Git

2. 克隆仓库
```bash
git clone https://github.com/yourusername/go-gin-websocket-msg.git
cd go-gin-websocket-msg
```

3. 安装依赖
```bash
go mod tidy
```

4. 运行应用
```bash
go run main.go
```

### 打包为可执行文件

#### 在Unix/Linux/macOS上打包

使用提供的shell脚本打包为所有平台:

```bash
chmod +x build.sh
./build.sh
```

#### 在Windows上打包

使用提供的批处理脚本打包为所有平台:

```bash
build.bat
```

构建好的可执行文件将位于`dist`目录中，包含以下文件:
- `chat-app-windows-amd64.zip` - Windows 64位版本（内存数据库模式）
- `chat-app-macos-amd64.tar.gz` - macOS Intel版本（内存数据库模式）
- `chat-app-linux-amd64.tar.gz` - Linux 64位版本（内存数据库模式）
- `chat-app-macos-amd64-with-sqlite.tar.gz` - macOS SQLite支持版本（仅在macOS上构建时）
- `chat-app-linux-amd64-with-sqlite.tar.gz` - Linux SQLite支持版本（仅在Linux上构建时）

### 运行已打包的应用

1. 解压对应平台的压缩包
2. 运行可执行文件:
   - Windows: 双击 `chat-app-windows-amd64.exe`
   - macOS: 打开终端，进入解压目录，执行 `./chat-app-macos-amd64`
   - Linux: 打开终端，进入解压目录，执行 `./chat-app-linux-amd64`

3. 打开浏览器访问 `http://localhost:8080`

## 注意事项

- 应用默认使用8080端口，如果该端口被占用，请修改`main.go`中的端口设置
- 无CGO支持时（跨平台版本），应用将自动使用内存数据库模式，应用关闭后数据会丢失
- 有CGO支持时（本地编译版本），数据存储在SQLite数据库中，文件位于应用同级目录的`chatroom.db`
- 聊天室不需要登录，使用IP地址作为用户标识，可以通过设置昵称来区分用户

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
- 发送消息应该看到自己的消息显示在右侧

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


