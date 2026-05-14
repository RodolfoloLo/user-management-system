package model

import (
	"log"
	"ums/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB // 全局数据库连接
/*
gorm.DB是一个非常庞大且复杂的结构体类型,它内部包含了:
1,底层的TCP数据库连接池
2,日志系统
3,读写锁状态
4,各种各样的回调函数

这是一个指针
Go中,只要你需要"全局唯一不复制","让别人能修改我","允许为空",就果断上指针
*/

func InitDB() {
	var err error
	// 连接数据库
	DB, err = gorm.Open(postgres.Open(config.Conf.Database.DSN), &gorm.Config{})

	if err != nil {
		log.Fatalf("数据库连接失败:", err)
	}

	log.Println("数据库连接成功！")

	// 自动创建/更新表结构 (如果数据库里没这个表，GORM会自动建)
	DB.AutoMigrate(&User{})
}
