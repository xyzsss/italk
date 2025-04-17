package models

import (
	"database/sql"
	"log"
	"os"
	
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// 初始化数据库连接
func InitDB() {
	// 检查数据库文件是否存在，如果不存在则创建
	if _, err := os.Stat("./chat.db"); os.IsNotExist(err) {
		file, err := os.Create("./chat.db")
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", "./chat.db")
	if err != nil {
		log.Fatal(err)
	}

	// 设置连接参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// 检查连接
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	DB = db
	
	// 创建表
	createTables()
}

// 创建数据库表
func createTables() {
	// 创建用户表
	userTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		ip TEXT NOT NULL,
		last_online TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建消息表
	messageTable := `CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		content TEXT,
		type TEXT,
		status INTEGER DEFAULT 0,
		file_name TEXT,
		file_size INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	// 执行创建表操作
	_, err := DB.Exec(userTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(messageTable)
	if err != nil {
		log.Fatal(err)
	}
	
	// 检查并添加新列
	updateTables()
}

// 更新表结构，添加新列
func updateTables() {
	// 尝试添加status列
	_, err := DB.Exec(`ALTER TABLE messages ADD COLUMN status INTEGER DEFAULT 0`)
	if err != nil {
		// 列已存在错误可以忽略
		log.Printf("添加status列: %v", err)
	}
	
	// 尝试添加file_name列
	_, err = DB.Exec(`ALTER TABLE messages ADD COLUMN file_name TEXT`)
	if err != nil {
		log.Printf("添加file_name列: %v", err)
	}
	
	// 尝试添加file_size列
	_, err = DB.Exec(`ALTER TABLE messages ADD COLUMN file_size INTEGER`)
	if err != nil {
		log.Printf("添加file_size列: %v", err)
	}
} 