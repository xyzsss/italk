package models

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

// 消息类型常量
const (
	MessageTypeText   = 0
	MessageTypeImage  = 1
	MessageTypeEmoji  = 2
	MessageTypeSystem = 3
	MessageTypeFile   = 4
)

// 消息状态常量
const (
	MessageStatusNormal   = 0
	MessageStatusRecalled = 1
)

// Message 表示聊天消息
type Message struct {
	ID        int64         `json:"id"`
	UserID    int64         `json:"user_id"`
	Username  sql.NullString `json:"-"`
	UsernameStr string       `json:"username"`
	Content   string        `json:"content"`
	Type      int           `json:"type"`
	Status    int           `json:"status"`
	FileName  sql.NullString `json:"-"`
	FileNameStr string      `json:"file_name"`
	FileSize  sql.NullInt64 `json:"-"`
	FileSizeVal int64       `json:"file_size"`
	CreatedAt time.Time     `json:"created_at"`
}

// GetMessages 获取最近的消息
func GetMessages(limit int) ([]*Message, error) {
	if UseMemoryMode {
		return getMessagesMemory(limit)
	}
	
	// SQLite模式
	query := `
		SELECT m.id, m.user_id, u.username, m.content, m.type, m.status, m.file_name, m.file_size, m.created_at
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		ORDER BY m.created_at ASC
		LIMIT ?
	`
	
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.Type,
			&msg.Status,
			&msg.FileName,
			&msg.FileSize,
			&msg.CreatedAt,
		)
		if err != nil {
			log.Printf("扫描消息行失败: %v", err)
			continue
		}
		
		// 设置用户友好字段
		if msg.Username.Valid {
			msg.UsernameStr = msg.Username.String
		}
		if msg.FileName.Valid {
			msg.FileNameStr = msg.FileName.String
		}
		if msg.FileSize.Valid {
			msg.FileSizeVal = msg.FileSize.Int64
		}
		
		messages = append(messages, &msg)
	}
	
	return messages, nil
}

// 内存模式下获取消息
func getMessagesMemory(limit int) ([]*Message, error) {
	messageMutex.RLock()
	defer messageMutex.RUnlock()
	
	// 收集所有消息
	messages := make([]*Message, 0, len(MessagesMap))
	for _, msg := range MessagesMap {
		messages = append(messages, msg)
	}
	
	// 按创建时间排序（从早到晚）
	for i := 0; i < len(messages); i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].CreatedAt.After(messages[j].CreatedAt) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}
	
	// 限制消息数量
	if len(messages) > limit {
		start := len(messages) - limit
		messages = messages[start:]
	}
	
	return messages, nil
}

// SearchMessages 搜索消息
func SearchMessages(query string) ([]*Message, error) {
	if UseMemoryMode {
		return searchMessagesMemory(query)
	}
	
	// SQLite模式
	sqlQuery := `
		SELECT m.id, m.user_id, u.username, m.content, m.type, m.status, m.file_name, m.file_size, m.created_at
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.content LIKE ? OR u.username LIKE ? OR m.file_name LIKE ?
		ORDER BY m.created_at DESC
		LIMIT 100
	`
	
	searchTerm := "%" + query + "%"
	rows, err := DB.Query(sqlQuery, searchTerm, searchTerm, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.Type,
			&msg.Status,
			&msg.FileName,
			&msg.FileSize,
			&msg.CreatedAt,
		)
		if err != nil {
			log.Printf("扫描消息行失败: %v", err)
			continue
		}
		
		// 设置用户友好字段
		if msg.Username.Valid {
			msg.UsernameStr = msg.Username.String
		}
		if msg.FileName.Valid {
			msg.FileNameStr = msg.FileName.String
		}
		if msg.FileSize.Valid {
			msg.FileSizeVal = msg.FileSize.Int64
		}
		
		messages = append(messages, &msg)
	}
	
	return messages, nil
}

// 内存模式下搜索消息
func searchMessagesMemory(query string) ([]*Message, error) {
	messageMutex.RLock()
	defer messageMutex.RUnlock()
	
	messages := make([]*Message, 0)
	
	// 在内存中搜索匹配的消息
	for _, msg := range MessagesMap {
		// 检查内容、用户名或文件名是否匹配
		if containsIgnoreCase(msg.Content, query) || 
			containsIgnoreCase(msg.UsernameStr, query) || 
			containsIgnoreCase(msg.FileNameStr, query) {
			messages = append(messages, msg)
		}
	}
	
	// 按时间排序（从新到旧）
	for i := 0; i < len(messages); i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].CreatedAt.Before(messages[j].CreatedAt) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}
	
	// 限制结果数量
	if len(messages) > 100 {
		messages = messages[:100]
	}
	
	return messages, nil
}

// 不区分大小写的子字符串检查
func containsIgnoreCase(s, substr string) bool {
	// 简单实现，不用引入strings包
	s, substr = toLower(s), toLower(substr)
	
	// 检查s中是否包含substr
	return indexOf(s, substr) >= 0
}

// 简单的转小写函数
func toLower(s string) string {
	bytes := []byte(s)
	for i, b := range bytes {
		if 'A' <= b && b <= 'Z' {
			bytes[i] = b + ('a' - 'A')
		}
	}
	return string(bytes)
}

// 简单的子字符串查找函数
func indexOf(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}

// CreateMessage 创建新消息
func CreateMessage(userID int64, content string, msgType int) (*Message, error) {
	if UseMemoryMode {
		return createMessageMemory(userID, content, msgType, "", 0)
	}
	
	// SQLite模式
	query := `INSERT INTO messages (user_id, content, type, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`
	result, err := DB.Exec(query, userID, content, msgType)
	if err != nil {
		return nil, err
	}
	
	msgID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// 获取插入的消息
	var msg Message
	query = `SELECT id, user_id, content, type, status, created_at FROM messages WHERE id = ?`
	err = DB.QueryRow(query, msgID).Scan(
		&msg.ID,
		&msg.UserID,
		&msg.Content,
		&msg.Type,
		&msg.Status,
		&msg.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	// 获取用户名
	var username sql.NullString
	query = `SELECT username FROM users WHERE id = ?`
	err = DB.QueryRow(query, userID).Scan(&username)
	if err == nil && username.Valid {
		msg.UsernameStr = username.String
	}
	
	return &msg, nil
}

// CreateFileMessage 创建文件消息
func CreateFileMessage(userID int64, content string, fileName string, fileSize int64) (*Message, error) {
	if UseMemoryMode {
		return createMessageMemory(userID, content, MessageTypeFile, fileName, fileSize)
	}
	
	// SQLite模式
	query := `INSERT INTO messages (user_id, content, type, file_name, file_size, created_at) 
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
	result, err := DB.Exec(query, userID, content, MessageTypeFile, fileName, fileSize)
	if err != nil {
		return nil, err
	}
	
	msgID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// 获取插入的消息
	var msg Message
	query = `SELECT id, user_id, content, type, status, file_name, file_size, created_at FROM messages WHERE id = ?`
	err = DB.QueryRow(query, msgID).Scan(
		&msg.ID,
		&msg.UserID,
		&msg.Content,
		&msg.Type,
		&msg.Status,
		&msg.FileName,
		&msg.FileSize,
		&msg.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	// 设置字段值
	if msg.FileName.Valid {
		msg.FileNameStr = msg.FileName.String
	}
	if msg.FileSize.Valid {
		msg.FileSizeVal = msg.FileSize.Int64
	}
	
	// 获取用户名
	var username sql.NullString
	query = `SELECT username FROM users WHERE id = ?`
	err = DB.QueryRow(query, userID).Scan(&username)
	if err == nil && username.Valid {
		msg.UsernameStr = username.String
	}
	
	return &msg, nil
}

// 内存模式下创建消息
func createMessageMemory(userID int64, content string, msgType int, fileName string, fileSize int64) (*Message, error) {
	messageMutex.Lock()
	defer messageMutex.Unlock()
	
	// 生成新消息ID
	LastMsgID++
	
	msg := &Message{
		ID:        LastMsgID,
		UserID:    userID,
		Content:   content,
		Type:      msgType,
		Status:    MessageStatusNormal,
		CreatedAt: time.Now(),
	}
	
	// 如果是文件消息，设置文件属性
	if msgType == MessageTypeFile && fileName != "" {
		msg.FileName.Valid = true
		msg.FileName.String = fileName
		msg.FileNameStr = fileName
		
		msg.FileSize.Valid = true
		msg.FileSize.Int64 = fileSize
		msg.FileSizeVal = fileSize
	}
	
	// 获取用户名
	usersMutex.RLock()
	if user, exists := UsersMap[userID]; exists && user.Username.Valid {
		msg.Username = user.Username
		msg.UsernameStr = user.Username.String
	}
	usersMutex.RUnlock()
	
	// 存储消息
	MessagesMap[msg.ID] = msg
	
	return msg, nil
}

// RecallMessage 撤回消息
func RecallMessage(messageID, userID int64) error {
	if UseMemoryMode {
		return recallMessageMemory(messageID, userID)
	}
	
	// SQLite模式
	// 首先获取消息信息
	var msg Message
	query := `SELECT id, user_id, created_at FROM messages WHERE id = ?`
	err := DB.QueryRow(query, messageID).Scan(&msg.ID, &msg.UserID, &msg.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("消息不存在")
		}
		return err
	}
	
	// 检查是否是消息的发送者
	if msg.UserID != userID {
		return errors.New("无权撤回他人消息")
	}
	
	// 检查消息是否在8小时内发送
	if time.Since(msg.CreatedAt) > 8*time.Hour {
		return errors.New("只能撤回8小时内的消息")
	}
	
	// 更新消息状态为已撤回
	query = `UPDATE messages SET status = ? WHERE id = ?`
	_, err = DB.Exec(query, MessageStatusRecalled, messageID)
	return err
}

// 内存模式下撤回消息
func recallMessageMemory(messageID, userID int64) error {
	messageMutex.Lock()
	defer messageMutex.Unlock()
	
	// 查找消息
	msg, exists := MessagesMap[messageID]
	if !exists {
		return errors.New("消息不存在")
	}
	
	// 检查是否是消息的发送者
	if msg.UserID != userID {
		return errors.New("无权撤回他人消息")
	}
	
	// 检查消息是否在8小时内发送
	if time.Since(msg.CreatedAt) > 8*time.Hour {
		return errors.New("只能撤回8小时内的消息")
	}
	
	// 更新消息状态为已撤回
	msg.Status = MessageStatusRecalled
	
	return nil
}

// GetMessageByID 根据ID获取消息
func GetMessageByID(messageID int64) (*Message, error) {
	if UseMemoryMode {
		return getMessageByIDMemory(messageID)
	}
	
	// SQLite模式
	var msg Message
	query := `
		SELECT m.id, m.user_id, u.username, m.content, m.type, m.status, m.file_name, m.file_size, m.created_at
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.id = ?
	`
	
	err := DB.QueryRow(query, messageID).Scan(
		&msg.ID,
		&msg.UserID,
		&msg.Username,
		&msg.Content,
		&msg.Type,
		&msg.Status,
		&msg.FileName,
		&msg.FileSize,
		&msg.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("消息不存在")
		}
		return nil, err
	}
	
	// 设置用户友好字段
	if msg.Username.Valid {
		msg.UsernameStr = msg.Username.String
	}
	if msg.FileName.Valid {
		msg.FileNameStr = msg.FileName.String
	}
	if msg.FileSize.Valid {
		msg.FileSizeVal = msg.FileSize.Int64
	}
	
	return &msg, nil
}

// 内存模式下根据ID获取消息
func getMessageByIDMemory(messageID int64) (*Message, error) {
	messageMutex.RLock()
	defer messageMutex.RUnlock()
	
	msg, exists := MessagesMap[messageID]
	if !exists {
		return nil, errors.New("消息不存在")
	}
	
	return msg, nil
}

// GetStatistics 获取聊天室统计信息
func GetStatistics() (map[string]interface{}, error) {
	stats := map[string]interface{}{}
	
	// 获取用户数量
	var userCount int
	err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return nil, err
	}
	stats["user_count"] = userCount
	
	// 获取消息数量
	var messageCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM messages").Scan(&messageCount)
	if err != nil {
		return nil, err
	}
	stats["message_count"] = messageCount
	
	// 获取在线用户
	onlineUsers, err := GetOnlineUsers()
	if err != nil {
		return nil, err
	}
	stats["online_users"] = onlineUsers
	
	// 获取最近的消息
	recentMessages, err := GetMessages(50)
	if err != nil {
		return nil, err
	}
	stats["recent_messages"] = recentMessages
	
	return stats, nil
} 