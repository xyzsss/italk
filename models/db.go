package models

import (
	"database/sql"
	"errors"
	"log"
	"sync"
	
	_ "github.com/mattn/go-sqlite3"
)

// DB 是全局数据库连接
var DB *sql.DB

// 内存数据库模式使用的变量
var (
	UsersMap     = make(map[int64]*User)
	MessagesMap  = make(map[int64]*Message)
	usersMutex   = &sync.RWMutex{}
	messageMutex = &sync.RWMutex{}
	LastUserID   int64 = 0
	LastMsgID    int64 = 0
	UseMemoryMode      = false
)

// 数据库错误
var (
	ErrNotImplemented = errors.New("功能在内存模式下未实现")
	ErrNoRows         = errors.New("未找到记录")
)

// InitDB 初始化数据库连接
func InitDB() {
	// 检查CGO是否启用，尝试使用SQLite
	UseMemoryMode = false
	var err error
	
	// 尝试连接SQLite文件数据库
	DB, err = sql.Open("sqlite3", "chat.db")
	
	if err != nil || checkDBConnection() != nil {
		log.Printf("SQLite数据库不可用: %v", err)
		log.Println("切换到内存数据模式...")
		UseMemoryMode = true
		
		// 初始化内存数据
		initMemoryData()
		return
	}
	
	// 创建表
	createTables()
	
	// 更新表结构
	updateTables()
	
	log.Println("数据库初始化完成，使用SQLite文件数据库")
}

// 检查数据库连接是否正常
func checkDBConnection() error {
	return DB.Ping()
}

// createTables 创建数据库表
func createTables() {
	// 创建用户表
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip TEXT NOT NULL,
			username TEXT,
			last_online TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("创建用户表失败: %v", err)
	}
	
	// 创建消息表
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			content TEXT NOT NULL,
			type INTEGER DEFAULT 0,
			status INTEGER DEFAULT 0,
			file_name TEXT,
			file_size INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)
	`)
	if err != nil {
		log.Printf("创建消息表失败: %v", err)
	}
}

// updateTables 更新表结构，添加新列
func updateTables() {
	// 添加消息状态列
	_, err := DB.Exec("ALTER TABLE messages ADD COLUMN status INTEGER DEFAULT 0")
	if err != nil {
		// 忽略错误，列可能已存在
		log.Printf("添加status列: %v", err)
	}
	
	// 添加文件名列
	_, err = DB.Exec("ALTER TABLE messages ADD COLUMN file_name TEXT")
	if err != nil {
		// 忽略错误，列可能已存在
		log.Printf("添加file_name列: %v", err)
	}
	
	// 添加文件大小列
	_, err = DB.Exec("ALTER TABLE messages ADD COLUMN file_size INTEGER")
	if err != nil {
		// 忽略错误，列可能已存在
		log.Printf("添加file_size列: %v", err)
	}
}

// 初始化内存数据
func initMemoryData() {
	log.Println("初始化内存数据模式...")
	// 清空现有数据
	UsersMap = make(map[int64]*User)
	MessagesMap = make(map[int64]*Message)
	LastUserID = 0
	LastMsgID = 0
} 