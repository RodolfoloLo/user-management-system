package main

import (
	"fmt"

	"ums/internal/config"
	"ums/internal/model"
	"ums/internal/router"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.InitConfig()
	model.InitDB()
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	router.SetupRouter(e)
	port := fmt.Sprintf(":%d", config.Conf.Server.Port)
	e.Logger.Fatal(e.Start(port))
}
