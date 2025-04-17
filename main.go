package main

import (
	"log"
	"net"
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/mikewang/go-gin-websocket-msg/controllers"
	"github.com/mikewang/go-gin-websocket-msg/models"
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

// 清理长时间未活跃用户
func cleanupInactiveUsers() {
	// 删除超过5分钟未活跃的用户
	err := models.CleanupInactiveUsers(5)
	if err != nil {
		log.Printf("清理不活跃用户失败: %v", err)
	}
}

func main() {
	// 初始化数据库
	models.InitDB()
	
	// 清理长时间未活跃的用户
	cleanupInactiveUsers()
	
	// 定期清理，每分钟执行一次
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			cleanupInactiveUsers()
		}
	}()
	
	// 初始化WebSocket消息处理函数
	controllers.InitMessageHandler()
	
	// 创建Gin引擎
	r := gin.Default()
	
	// 设置静态文件服务
	r.Static("/static", "./static")
	
	// 加载HTML模板
	r.LoadHTMLGlob("templates/*")
	
	// 根路由 - 聊天室页面
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "局域网聊天室",
		})
	})
	
	// WebSocket路由
	r.GET("/ws", controllers.HandleWebSocket)
	
	// API路由
	api := r.Group("/api")
	{
		// 获取历史消息
		api.GET("/messages", controllers.GetMessages)
		
		// 搜索历史消息
		api.GET("/messages/search", controllers.SearchMessages)
		
		// 获取在线用户列表
		api.GET("/users/online", controllers.GetOnlineUsers)
		
		// 获取聊天室统计信息
		api.GET("/stats", controllers.GetStatistics)
	}
	
	// 获取本机IP
	localIP, err := getLocalIP()
	if err != nil {
		log.Printf("获取本机IP失败: %v", err)
		localIP = "127.0.0.1"
	}
	
	// 打印服务信息
	log.Printf("聊天服务运行在: http://%s:8080", localIP)
	
	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
} 