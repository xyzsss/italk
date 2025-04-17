package models

import (
	"database/sql"
	"fmt"
	"time"
)

// 消息类型常量
const (
	MessageTypeText   = "text"
	MessageTypeImage  = "image"
	MessageTypeEmoji  = "emoji"
	MessageTypeSystem = "system"
	MessageTypeFile   = "file"  // 新增文件类型
)

// 消息状态常量
const (
	MessageStatusNormal    = 0  // 正常
	MessageStatusRecalled  = 1  // 已撤回
)

// Message 表示聊天消息
type Message struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Status    int       `json:"status"`     // 消息状态
	CreatedAt time.Time `json:"created_at"`
	
	// 附加信息，用于前端显示
	Username string `json:"username,omitempty"`
	IP       string `json:"ip,omitempty"`
	FileName string `json:"file_name,omitempty"` // 文件名
	FileSize int64  `json:"file_size,omitempty"` // 文件大小
}

// CreateMessage 创建新消息
func CreateMessage(userID int64, content, msgType string) (*Message, error) {
	msg := &Message{
		UserID:  userID,
		Content: content,
		Type:    msgType,
		Status:  MessageStatusNormal, // 默认消息状态为正常
	}

	// 使用两步操作替代RETURNING子句
	query := `INSERT INTO messages (user_id, content, type, status) VALUES (?, ?, ?, ?)`
	result, err := DB.Exec(query, userID, content, msgType, MessageStatusNormal)
	if err != nil {
		return nil, err
	}
	
	// 获取新插入行的ID
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	msg.ID = id
	
	// 再次查询获取完整消息信息
	query = `SELECT id, created_at FROM messages WHERE id = ?`
	err = DB.QueryRow(query, id).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// RecallMessage 撤回消息
func RecallMessage(msgID int64, userID int64) error {
	// 验证消息属于该用户且在8小时以内
	var createdAt time.Time
	err := DB.QueryRow(`SELECT created_at FROM messages WHERE id = ? AND user_id = ?`, 
		msgID, userID).Scan(&createdAt)
	if err != nil {
		return err
	}
	
	// 检查消息是否在8小时内
	if time.Since(createdAt).Hours() > 8 {
		return fmt.Errorf("消息已超过8小时，无法撤回")
	}
	
	// 更新消息状态为已撤回
	_, err = DB.Exec(`UPDATE messages SET status = ? WHERE id = ?`, 
		MessageStatusRecalled, msgID)
	return err
}

// GetMessageByID 根据ID获取消息
func GetMessageByID(msgID int64) (*Message, error) {
	msg := &Message{}
	query := `SELECT m.id, m.user_id, m.content, m.type, m.status, m.created_at, 
			  u.username, u.ip
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.id = ?`
		
	var username sql.NullString
	err := DB.QueryRow(query, msgID).Scan(
		&msg.ID, 
		&msg.UserID, 
		&msg.Content, 
		&msg.Type, 
		&msg.Status,
		&msg.CreatedAt,
		&username,
		&msg.IP,
	)
	if err != nil {
		return nil, err
	}
	
	// 转换可空用户名
	if username.Valid {
		msg.Username = username.String
	} else {
		msg.Username = ""
	}
	
	return msg, nil
}

// GetMessages 获取最近的消息
func GetMessages(limit int) ([]*Message, error) {
	messages := []*Message{}
	
	query := `SELECT m.id, m.user_id, m.content, m.type, m.status, m.created_at, 
			  u.username, u.ip, m.file_name, m.file_size
		FROM messages m
		JOIN users u ON m.user_id = u.id
		ORDER BY m.created_at DESC
		LIMIT ?`
		
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		msg := &Message{}
		var username sql.NullString
		var fileName sql.NullString
		var fileSize sql.NullInt64
		err := rows.Scan(
			&msg.ID, 
			&msg.UserID, 
			&msg.Content, 
			&msg.Type, 
			&msg.Status,
			&msg.CreatedAt,
			&username,
			&msg.IP,
			&fileName,
			&fileSize,
		)
		if err != nil {
			return nil, err
		}
		
		// 转换可空用户名
		if username.Valid {
			msg.Username = username.String
		} else {
			msg.Username = ""
		}
		
		// 转换可空文件信息
		if fileName.Valid {
			msg.FileName = fileName.String
		}
		if fileSize.Valid {
			msg.FileSize = fileSize.Int64
		}
		
		messages = append(messages, msg)
	}
	
	// 反转消息顺序，使最早的消息在前
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, nil
}

// SearchMessages 搜索消息
func SearchMessages(query string) ([]*Message, error) {
	messages := []*Message{}
	
	sqlQuery := `SELECT m.id, m.user_id, m.content, m.type, m.status, m.created_at, 
			  u.username, u.ip, m.file_name, m.file_size
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.content LIKE ? OR u.username LIKE ? OR u.ip LIKE ? OR m.type LIKE ? OR m.file_name LIKE ?
		ORDER BY m.created_at DESC
		LIMIT 100`
		
	searchParam := "%" + query + "%"
	rows, err := DB.Query(sqlQuery, searchParam, searchParam, searchParam, searchParam, searchParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		msg := &Message{}
		var username sql.NullString
		var fileName sql.NullString
		var fileSize sql.NullInt64
		err := rows.Scan(
			&msg.ID, 
			&msg.UserID, 
			&msg.Content, 
			&msg.Type, 
			&msg.Status,
			&msg.CreatedAt,
			&username,
			&msg.IP,
			&fileName,
			&fileSize,
		)
		if err != nil {
			return nil, err
		}
		
		// 转换可空用户名
		if username.Valid {
			msg.Username = username.String
		} else {
			msg.Username = ""
		}
		
		// 转换可空文件信息
		if fileName.Valid {
			msg.FileName = fileName.String
		}
		if fileSize.Valid {
			msg.FileSize = fileSize.Int64
		}
		
		messages = append(messages, msg)
	}
	
	return messages, nil
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

// CreateFileMessage 创建文件消息
func CreateFileMessage(userID int64, content, fileName string, fileSize int64) (*Message, error) {
	msg := &Message{
		UserID:   userID,
		Content:  content,
		Type:     MessageTypeFile,
		Status:   MessageStatusNormal,
		FileName: fileName,
		FileSize: fileSize,
	}

	query := `INSERT INTO messages (user_id, content, type, status, file_name, file_size) 
		VALUES (?, ?, ?, ?, ?, ?)`
	result, err := DB.Exec(query, userID, content, MessageTypeFile, MessageStatusNormal, fileName, fileSize)
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	msg.ID = id
	
	query = `SELECT id, created_at FROM messages WHERE id = ?`
	err = DB.QueryRow(query, id).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	return msg, nil
} 