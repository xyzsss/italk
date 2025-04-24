package models

import (
	"database/sql"
	"log"
	"time"
)

// User 表示聊天用户
type User struct {
	ID           int64        `json:"id"`
	IP           string       `json:"ip"`
	Username     sql.NullString `json:"-"`
	UsernameStr  string       `json:"username"`
	LastOnline   time.Time    `json:"last_online"`
}

// GetUserByIP 根据IP地址获取用户
func GetUserByIP(ip string) (*User, error) {
	if UseMemoryMode {
		return getUserByIPMemory(ip)
	}
	
	// SQLite模式
	var user User
	query := `SELECT id, ip, username, last_online FROM users WHERE ip = ?`
	err := DB.QueryRow(query, ip).Scan(&user.ID, &user.IP, &user.Username, &user.LastOnline)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在，创建新用户
			return CreateUser(ip, "")
		}
		return nil, err
	}
	
	// 更新最后在线时间
	_, err = DB.Exec(`UPDATE users SET last_online = CURRENT_TIMESTAMP WHERE id = ?`, user.ID)
	if err != nil {
		log.Printf("更新用户最后在线时间失败: %v", err)
	}
	
	user.UsernameStr = user.Username.String
	
	return &user, nil
}

// 内存模式下根据IP获取用户
func getUserByIPMemory(ip string) (*User, error) {
	usersMutex.RLock()
	defer usersMutex.RUnlock()
	
	// 查找IP匹配的用户
	for _, user := range UsersMap {
		if user.IP == ip {
			// 更新最后在线时间
			user.LastOnline = time.Now()
			
			return user, nil
		}
	}
	
	// 用户不存在，创建新用户
	return CreateUser(ip, "")
}

// CreateUser 创建新用户
func CreateUser(ip, username string) (*User, error) {
	if UseMemoryMode {
		return createUserMemory(ip, username)
	}
	
	// SQLite模式
	var user User
	user.IP = ip
	
	if username != "" {
		user.Username.Valid = true
		user.Username.String = username
		user.UsernameStr = username
	}
	
	// 插入数据库
	query := `INSERT INTO users (ip, username, last_online) VALUES (?, ?, CURRENT_TIMESTAMP)`
	result, err := DB.Exec(query, ip, user.Username)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		return nil, err
	}
	
	// 获取插入ID
	user.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	user.LastOnline = time.Now()
	
	return &user, nil
}

// 内存模式下创建新用户
func createUserMemory(ip, username string) (*User, error) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	
	// 生成新ID
	LastUserID++
	
	user := &User{
		ID:         LastUserID,
		IP:         ip,
		LastOnline: time.Now(),
	}
	
	if username != "" {
		user.Username.Valid = true
		user.Username.String = username
		user.UsernameStr = username
	}
	
	// 存储到内存
	UsersMap[user.ID] = user
	
	return user, nil
}

// UpdateUsername 更新用户名
func UpdateUsername(userID int64, username string) error {
	if UseMemoryMode {
		return updateUsernameMemory(userID, username)
	}
	
	// SQLite模式
	if username == "" {
		return nil // 空用户名不处理
	}
	
	_, err := DB.Exec(`UPDATE users SET username = ? WHERE id = ?`, username, userID)
	return err
}

// 内存模式下更新用户名
func updateUsernameMemory(userID int64, username string) error {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	
	if username == "" {
		return nil // 空用户名不处理
	}
	
	user, exists := UsersMap[userID]
	if !exists {
		return ErrNoRows
	}
	
	user.Username.Valid = true
	user.Username.String = username
	user.UsernameStr = username
	
	return nil
}

// GetOnlineUsers 获取最近30秒内在线的用户
func GetOnlineUsers() ([]*User, error) {
	if UseMemoryMode {
		return getOnlineUsersMemory()
	}
	
	// SQLite模式
	query := `SELECT id, ip, username, last_online FROM users 
		WHERE last_online > datetime('now', '-30 seconds') 
		ORDER BY last_online DESC`
	
	rows, err := DB.Query(query)
	if err != nil {
		log.Printf("获取在线用户失败: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.IP, &user.Username, &user.LastOnline)
		if err != nil {
			log.Printf("扫描用户行失败: %v", err)
			continue
		}
		
		user.UsernameStr = user.Username.String
		users = append(users, &user)
	}
	
	log.Printf("当前在线用户数: %d", len(users))
	return users, nil
}

// 内存模式下获取在线用户
func getOnlineUsersMemory() ([]*User, error) {
	usersMutex.RLock()
	defer usersMutex.RUnlock()
	
	var users []*User
	threshold := time.Now().Add(-30 * time.Second)
	
	for _, user := range UsersMap {
		if user.LastOnline.After(threshold) {
			users = append(users, user)
		}
	}
	
	log.Printf("内存模式 - 当前在线用户数: %d", len(users))
	return users, nil
}

// CleanupInactiveUsers 清理超过1分钟未活动的用户
func CleanupInactiveUsers() error {
	if UseMemoryMode {
		return cleanupInactiveUsersMemory()
	}

	// SQLite模式
	query := `DELETE FROM users WHERE last_online < datetime('now', '-1 minute')`
	result, err := DB.Exec(query)
	if err != nil {
		log.Printf("清理不活跃用户失败: %v", err)
		return err
	}

	affected, _ := result.RowsAffected()
	if affected > 0 {
		log.Printf("已清理 %d 个不活跃用户", affected)
	}
	return nil
}

// 内存模式下清理不活跃用户
func cleanupInactiveUsersMemory() error {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	threshold := time.Now().Add(-1 * time.Minute)
	initialCount := len(UsersMap)

	for ip, user := range UsersMap {
		if user.LastOnline.Before(threshold) {
			delete(UsersMap, ip)
		}
	}

	cleaned := initialCount - len(UsersMap)
	if cleaned > 0 {
		log.Printf("内存模式 - 已清理 %d 个不活跃用户", cleaned)
	}
	return nil
}

// 内存模式下检查用户是否有消息
func userHasMessages(userID int64) bool {
	messageMutex.RLock()
	defer messageMutex.RUnlock()
	
	for _, msg := range MessagesMap {
		if msg.UserID == userID {
			return true
		}
	}
	
	return false
}

// UpdateLastOnline 更新用户的最后在线时间
func UpdateLastOnline(userID int64) error {
	if UseMemoryMode {
		return updateLastOnlineMemory(userID)
	}
	
	// SQLite模式
	_, err := DB.Exec(`UPDATE users SET last_online = CURRENT_TIMESTAMP WHERE id = ?`, userID)
	if err != nil {
		log.Printf("更新用户最后在线时间失败: %v", err)
	}
	return err
}

// 内存模式下更新最后在线时间
func updateLastOnlineMemory(userID int64) error {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	
	user, exists := UsersMap[userID]
	if !exists {
		return ErrNoRows
	}
	
	user.LastOnline = time.Now()
	return nil
} 