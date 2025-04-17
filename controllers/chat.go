package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mikewang/go-gin-websocket-msg/models"
	"github.com/mikewang/go-gin-websocket-msg/utils"
)

// Hub 是WebSocket hub的实例
var Hub *utils.Hub

// 活跃用户映射表，用于追踪实际在线的WebSocket连接
var (
	activeUsers = make(map[int64]bool)
	usersMutex  = &sync.Mutex{}
)

// 添加活跃用户
func addActiveUser(userID int64) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	activeUsers[userID] = true
}

// 移除活跃用户
func removeActiveUser(userID int64) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	delete(activeUsers, userID)
}

// 检查用户是否活跃
func isUserActive(userID int64) bool {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	return activeUsers[userID]
}

// 获取活跃用户数量
func getActiveUserCount() int {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	return len(activeUsers)
}

// 初始化Hub
func init() {
	Hub = utils.NewHub()
	go Hub.Run()
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(c *gin.Context) {
	// 获取客户端IP
	ip := c.ClientIP()
	
	// 尝试获取已有用户
	user, err := models.GetUserByIP(ip)
	if err != nil {
		// 创建新用户
		user, err = models.CreateUser(ip)
		if err != nil {
			log.Printf("创建用户失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
			return
		}
		
		// 广播用户进入聊天室的系统消息
		systemMsg := &utils.Message{
			Type:    utils.MessageTypeSystem,
			Content: ip + " 进入了聊天室",
		}
		Hub.BroadcastMessage(systemMsg)
		
		// 保存系统消息到数据库
		_, err = models.CreateMessage(user.ID, systemMsg.Content, models.MessageTypeSystem)
		if err != nil {
			log.Printf("保存系统消息失败: %v", err)
		}
	} else {
		// 更新用户最后在线时间
		err = models.UpdateLastOnline(user.ID)
		if err != nil {
			log.Printf("更新用户最后在线时间失败: %v", err)
		}
	}
	
	// 升级HTTP连接为WebSocket连接
	// 创建自定义的ServeWs处理函数，以便我们处理消息
	conn, err := utils.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	
	client := &utils.Client{
		ID:   user.ID,
		IP:   ip,
		Hub:  Hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}
	client.Hub.Register <- client
	
	// 添加到活跃用户列表
	addActiveUser(user.ID)
	
	// 启动goroutine来处理WebSocket连接
	go handleWritePump(client)
	go handleReadPump(client)
}

// 处理WebSocket写入操作
func handleWritePump(client *utils.Client) {
	defer func() {
		client.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				// 通道已关闭
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// 添加队列中的所有消息
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// 处理WebSocket读取操作
func handleReadPump(client *utils.Client) {
	defer func() {
		// 用户断开连接时从活跃用户列表移除
		removeActiveUser(client.ID)
		
		// 用户断开连接
		systemMsg := &utils.Message{
			Type:    utils.MessageTypeSystem,
			Content: client.IP + " 离开了聊天室",
		}
		Hub.BroadcastMessage(systemMsg)
		
		// 保存系统消息到数据库
		_, err := models.CreateMessage(client.ID, systemMsg.Content, models.MessageTypeSystem)
		if err != nil {
			log.Printf("保存系统消息失败: %v", err)
		}
		
		// 更新用户最后在线时间
		models.UpdateLastOnline(client.ID)
		
		client.Hub.Unregister <- client
		client.Conn.Close()
	}()
	
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("错误: %v", err)
			}
			break
		}
		
		var msg utils.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("错误: 解析消息失败: %v", err)
			continue
		}
		
		// 设置消息发送者信息
		msg.UserID = client.ID
		msg.IP = client.IP
		
		// 处理消息
		HandleMessage(client, &msg)
	}
}

// 处理WebSocket消息
func HandleMessage(client *utils.Client, msg *utils.Message) {
	// 根据消息类型处理消息
	switch msg.Type {
	case utils.MessageTypeText, utils.MessageTypeImage, utils.MessageTypeEmoji:
		// 处理文本、图片和表情消息
		handleChatMessage(client, msg)
	case utils.MessageTypeUser:
		// 处理用户信息更新（如更改用户名）
		handleUserUpdate(client, msg)
	case utils.MessageTypeFile:
		// 处理文件消息
		handleFileMessage(client, msg)
	case utils.MessageTypeRecall:
		// 处理消息撤回
		handleRecallMessage(client, msg)
	}
}

// 处理聊天消息（文本、图片和表情）
func handleChatMessage(client *utils.Client, msg *utils.Message) {
	// 获取用户信息
	user, err := models.GetUserByIP(client.IP)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		return
	}
	
	// 设置消息发送者信息
	msg.UserID = user.ID
	msg.Username = user.UsernameStr
	
	// 保存消息到数据库
	var msgType string
	switch msg.Type {
	case utils.MessageTypeText:
		msgType = models.MessageTypeText
	case utils.MessageTypeImage:
		msgType = models.MessageTypeImage
	case utils.MessageTypeEmoji:
		msgType = models.MessageTypeEmoji
	}
	
	dbMsg, err := models.CreateMessage(user.ID, msg.Content, msgType)
	if err != nil {
		log.Printf("保存消息失败: %v", err)
	} else {
		// 设置消息ID，便于后续撤回
		msg.MessageID = dbMsg.ID
	}
	
	// 广播消息给所有客户端
	Hub.BroadcastMessage(msg)
}

// 处理文件消息
func handleFileMessage(client *utils.Client, msg *utils.Message) {
	// 获取用户信息
	user, err := models.GetUserByIP(client.IP)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		return
	}
	
	// 设置消息发送者信息
	msg.UserID = user.ID
	msg.Username = user.UsernameStr
	
	// 保存消息到数据库
	dbMsg, err := models.CreateFileMessage(user.ID, msg.Content, msg.FileName, msg.FileSize)
	if err != nil {
		log.Printf("保存文件消息失败: %v", err)
	} else {
		// 设置消息ID，便于后续撤回
		msg.MessageID = dbMsg.ID
	}
	
	// 广播消息给所有客户端
	Hub.BroadcastMessage(msg)
}

// 处理消息撤回
func handleRecallMessage(client *utils.Client, msg *utils.Message) {
	if msg.MessageID == 0 {
		return
	}
	
	// 撤回消息
	err := models.RecallMessage(msg.MessageID, client.ID)
	if err != nil {
		log.Printf("撤回消息失败: %v", err)
		return
	}
	
	// 获取原消息信息，确认消息可以被撤回
	_, err = models.GetMessageByID(msg.MessageID)
	if err != nil {
		log.Printf("获取原消息失败: %v", err)
		return
	}
	
	// 创建撤回通知消息
	recallNotice := &utils.Message{
		Type:      utils.MessageTypeRecall,
		MessageID: msg.MessageID,
		UserID:    client.ID,
		Username:  client.IP, // 使用IP作为默认用户名
	}
	
	// 如果能获取到用户信息，则使用用户名
	user, err := models.GetUserByIP(client.IP)
	if err == nil && user.UsernameStr != "" {
		recallNotice.Username = user.UsernameStr
	}
	
	// 广播撤回通知给所有客户端
	Hub.BroadcastMessage(recallNotice)
	
	// 记录系统消息
	systemMsg := fmt.Sprintf("%s 撤回了一条消息", recallNotice.Username)
	_, err = models.CreateMessage(client.ID, systemMsg, models.MessageTypeSystem)
	if err != nil {
		log.Printf("保存撤回系统消息失败: %v", err)
	}
}

// 处理用户信息更新
func handleUserUpdate(client *utils.Client, msg *utils.Message) {
	if msg.Username == "" {
		return
	}
	
	// 更新用户名
	err := models.UpdateUsername(client.ID, msg.Username)
	if err != nil {
		log.Printf("更新用户名失败: %v", err)
		return
	}
	
	// 广播用户名更新的系统消息
	systemMsg := &utils.Message{
		Type:    utils.MessageTypeSystem,
		Content: client.IP + " 将昵称修改为 " + msg.Username,
	}
	Hub.BroadcastMessage(systemMsg)
	
	// 保存系统消息到数据库
	_, err = models.CreateMessage(client.ID, systemMsg.Content, models.MessageTypeSystem)
	if err != nil {
		log.Printf("保存系统消息失败: %v", err)
	}
}

// GetMessages 获取聊天历史消息
func GetMessages(c *gin.Context) {
	limit := 50
	messages, err := models.GetMessages(limit)
	if err != nil {
		log.Printf("获取消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息失败"})
		return
	}
	
	c.JSON(http.StatusOK, messages)
}

// SearchMessages 搜索聊天历史消息
func SearchMessages(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键字不能为空"})
		return
	}
	
	messages, err := models.SearchMessages(query)
	if err != nil {
		log.Printf("搜索消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索消息失败"})
		return
	}
	
	c.JSON(http.StatusOK, messages)
}

// GetOnlineUsers 获取在线用户列表
func GetOnlineUsers(c *gin.Context) {
	users, err := models.GetOnlineUsers()
	if err != nil {
		log.Printf("获取在线用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取在线用户失败"})
		return
	}
	
	// 过滤非真正活跃的用户
	activeUsersList := make([]*models.User, 0)
	for _, user := range users {
		if isUserActive(user.ID) {
			activeUsersList = append(activeUsersList, user)
		}
	}
	
	c.JSON(http.StatusOK, activeUsersList)
}

// GetStatistics 获取聊天室统计信息
func GetStatistics(c *gin.Context) {
	stats, err := models.GetStatistics()
	if err != nil {
		log.Printf("获取聊天室统计信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取聊天室统计信息失败"})
		return
	}
	
	// 更新活跃用户数
	stats["active_user_count"] = getActiveUserCount()
	
	// 过滤真正在线的用户
	if onlineUsers, ok := stats["online_users"].([]*models.User); ok {
		activeUsersList := make([]*models.User, 0)
		for _, user := range onlineUsers {
			if isUserActive(user.ID) {
				activeUsersList = append(activeUsersList, user)
			}
		}
		stats["online_users"] = activeUsersList
	}
	
	c.JSON(http.StatusOK, stats)
}

// InitMessageHandler 初始化消息处理函数
func InitMessageHandler() {
	// 因为无法直接替换Client的方法，我们会在ServeWs中手动处理消息
	// 这里不需要特殊处理
} 