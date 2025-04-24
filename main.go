package main

import (
	"fmt"
	"log"
	"net"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/mikewang/go-gin-websocket-msg/controllers"
	"github.com/mikewang/go-gin-websocket-msg/models"
	"github.com/mikewang/go-gin-websocket-msg/utils"
)

// 应用版本和构建时间，将在编译时通过ldflags注入
var (
	Version   = "dev"
	BuildTime = "unknown"
	ChatTitle = "局域网聊天室" // 聊天室名称配置
)

// 获取本机IP地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	
	return "127.0.0.1", nil
}

func main() {
	// 输出应用版本信息
	fmt.Printf("聊天室应用 版本: %s (构建时间: %s)\n", Version, BuildTime)
	
	// 初始化数据库
	models.InitDB()
	
	// 获取本机IP，显示访问地址
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("获取本机IP失败: %v", err)
		localIP = "127.0.0.1"
	}
	fmt.Printf("请通过浏览器访问: http://%s:8081\n", localIP)
	
	// 设置路由
	r := gin.Default()
	
	// 静态文件
	r.Static("/static", "./static")
	
	// HTML 模板
	r.LoadHTMLGlob("templates/*")
	
	// 聊天室首页
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title":     ChatTitle,
			"chatTitle": ChatTitle,
		})
	})
	
	// WebSocket 路由
	r.GET("/ws", controllers.HandleWebSocket)
	
	// API 路由
	r.GET("/api/messages", controllers.GetMessages)
	r.GET("/api/messages/search", controllers.SearchMessages)
	r.GET("/api/users/online", controllers.GetOnlineUsers)
	r.GET("/api/statistics", controllers.GetStatistics)
	
	// 更新聊天室标题
	r.POST("/api/title", func(c *gin.Context) {
		var req struct {
			Title string `json:"title" binding:"required"`
		}
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "标题不能为空"})
			return
		}
		
		// 更新标题
		ChatTitle = req.Title
		
		// 广播标题更新消息
		systemMsg := &utils.Message{
			Type:    utils.MessageTypeSystem,
			Content: "聊天室名称已更新为：" + ChatTitle,
		}
		controllers.Hub.BroadcastMessage(systemMsg)
		
		c.JSON(200, gin.H{
			"success": true,
			"title":   ChatTitle,
		})
	})
	
	// 启动定时清理任务
	go cleanupInactiveUsers()
	
	// 启动服务器
	fmt.Println("启动服务器，监听端口 8081...")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}

// 定时清理不活跃用户
func cleanupInactiveUsers() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		err := models.CleanupInactiveUsers()
		if err != nil {
			log.Printf("清理不活跃用户失败: %v", err)
		}
	}
} 