package utils

import "github.com/labstack/echo/v4"

type Response struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 消息
	Data    interface{} `json:"data"`    // 数据
}

func Success(c echo.Context, data interface{}) error {
	return c.JSON(200, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func Error(c echo.Context, code int, message string) error {
	return c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
