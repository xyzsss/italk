package models

import (
	"database/sql"
	"time"
)

// User 表示聊天室用户
type User struct {
	ID         int64          `json:"id"`
	Username   sql.NullString `json:"-"` // 使用sql.NullString处理NULL值
	UsernameStr string        `json:"username"` // 用于JSON序列化
	IP         string         `json:"ip"`
	LastOnline time.Time      `json:"last_online"`
}

// CreateUser 创建新用户
func CreateUser(ip string) (*User, error) {
	user := &User{
		IP: ip,
	}

	// 使用两步操作替代RETURNING子句
	query := `INSERT INTO users (ip) VALUES (?)`
	result, err := DB.Exec(query, ip)
	if err != nil {
		return nil, err
	}
	
	// 获取新插入行的ID
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	user.ID = id
	
	// 再次查询获取完整用户信息
	query = `SELECT id, username, ip, last_online FROM users WHERE id = ?`
	err = DB.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.IP, &user.LastOnline)
	if err != nil {
		return nil, err
	}

	// 转换可空用户名为普通字符串
	if user.Username.Valid {
		user.UsernameStr = user.Username.String
	} else {
		user.UsernameStr = ""
	}

	return user, nil
}

// GetUserByIP 通过IP获取用户
func GetUserByIP(ip string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, ip, last_online FROM users WHERE ip = ? ORDER BY last_online DESC LIMIT 1`
	err := DB.QueryRow(query, ip).Scan(&user.ID, &user.Username, &user.IP, &user.LastOnline)
	if err != nil {
		return nil, err
	}

	// 转换可空用户名为普通字符串
	if user.Username.Valid {
		user.UsernameStr = user.Username.String
	} else {
		user.UsernameStr = ""
	}

	return user, nil
}

// UpdateUsername 更新用户名
func UpdateUsername(id int64, username string) error {
	query := `UPDATE users SET username = ? WHERE id = ?`
	_, err := DB.Exec(query, username, id)
	return err
}

// UpdateLastOnline 更新用户最后在线时间
func UpdateLastOnline(id int64) error {
	query := `UPDATE users SET last_online = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := DB.Exec(query, id)
	return err
}

// GetOnlineUsers 获取在线用户列表（在过去1分钟内活跃的用户）
func GetOnlineUsers() ([]*User, error) {
	users := []*User{}
	
	query := `SELECT id, username, ip, last_online FROM users 
		WHERE last_online > datetime('now', '-1 minutes') 
		ORDER BY last_online DESC`
		
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.IP, &user.LastOnline)
		if err != nil {
			return nil, err
		}
		
		// 转换可空用户名为普通字符串
		if user.Username.Valid {
			user.UsernameStr = user.Username.String
		} else {
			user.UsernameStr = ""
		}
		
		users = append(users, user)
	}
	
	return users, nil
}

// CleanupInactiveUsers 清理指定分钟数内未活跃的用户
func CleanupInactiveUsers(minutes int) error {
	query := `DELETE FROM users WHERE last_online < datetime('now', '-? minutes')`
	_, err := DB.Exec(query, minutes)
	return err
} 