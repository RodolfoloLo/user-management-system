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
		log.Fatalf("数据库连接失败:%v", err)
	}

	log.Println("数据库连接成功！")

	// 自动创建/更新表结构 (如果数据库里没这个表，GORM会自动建)
	DB.AutoMigrate(&User{})
}

/*
对于log.Fatalf和panic的区别:
1. log.Fatalf会先打印一条日志,然后调用os.Exit(1)退出程序,而panic会引发一个运行时错误,如果没有被recover捕获,程序也会崩溃退出。
2. log.Fatalf不会执行defer语句,而panic会执行defer语句。(defer语句在函数返回前执行,无论是正常返回还是因为panic引起的异常返回都会执行)
3. log.Fatalf适用于记录错误日志并退出程序的场景,而panic适用于发生不可恢复的错误时引发异常的场景。
 在生产环境中，log.Fatalf 报错和 panic 哪个更正规？
要分运行阶段来看：
初始化阶段（启动时读取配置、连接数据库）：推荐 log.Fatalf。
启动时如果连接不上数据库或找不到配置文件，程序根本没有继续运行的意义（这叫 Fail-Fast 原则）。
panic() 会不仅退出程序，还会打印几百行极其难看的堆栈调用链（Stack Trace）；而你用的 log.Fatalf 会优雅地打印出“数据库连接失败”，然后以状态码 1 退出进程。对于明确的环境依赖错误，log.Fatalf 更加干净、专业。
服务运行阶段（处理 HTTP 请求时）：绝对不能主动用这俩。
在处理业务时，遇到错只能向上返回普通 error，然后使用普通的 log.Println 或 log.Error 记录一条日志，然后给前端返回错误 JSON。绝对不能去中断进程。
*/
