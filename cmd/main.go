package main

import (
	"fmt"
	"net/http"
	"ums/internal/config" // 注意这里是你自己的模块名

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. 初始化配置 (相当于导入 pydantic 实例)
	config.InitConfig()

	// 2. 初始化 Echo 框架 (相当于 app = FastAPI())
	e := echo.New()

	// 3. 全局中间件
	e.Use(middleware.Logger())  // 记录所有请求的日志
	e.Use(middleware.Recover()) // 处理 panic，防止程序崩溃 (相当于 FastAPI 默认防崩)

	// 4. 定义简单的测试路由
	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "pong",
		})
	})

	// 5. 启动服务
	port := fmt.Sprintf(":%d", config.Conf.Server.Port) // 拼接如 ":8080"
	e.Logger.Fatal(e.Start(port))                       // 启动，如果出错就打印并退出
}
