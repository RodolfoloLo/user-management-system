package router

import (
	"ums/internal/config"
	"ums/internal/controller"
	"ums/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func SetupRouter(e *echo.Echo) {
	api := e.Group("/api")

	// 公开路由 (不拦)
	api.POST("/register", controller.Register)
	api.POST("/login", controller.Login)

	// 私密路由 (JWT 拦截)
	authGroup := api.Group("/v1")

	// 配置 JWT 中间件：关键在于告诉 echojwt 我们使用了结构体 JWTClaims！
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(utils.JWTClaims) // 让框架解析到我们的自定义Struct中
		},
		SigningKey: []byte(config.Conf.Jwt.Secret),
	}
	authGroup.Use(echojwt.WithConfig(config))

	// 用户操作
	authGroup.GET("/user/me", controller.GetUserInfo)
	authGroup.PUT("/user/me", controller.UpdateUser)

	// 管理员操作
	authGroup.DELETE("/users/:id", controller.AdminDeleteUser)
}

/* 关于 echo-jwt 中间件的使用
中间件就像一个保安,站在/api/v1这个门口
当前端发送带有 Header: Authorization Bearer xxx 密文的请求试图访问某个接口时，在走到任何具体的 Controller 函数（比如 GetUserInfo）之前，这个特种兵保安会先把它拦下！
保安自己拿到一长串神秘火星文之后,会自己用密钥去解密,自己根据我们之前定义的 JWTClaims 结构体去解析出里面的用户ID和管理员信息。
参考自echo官方文档:
JWT provides a JSON Web Token (JWT) authentication middleware. Echo JWT middleware is located at https://github.com/labstack/echo-jwt

Basic middleware behavior:

For valid token, it sets the user in context and calls next handler.
For invalid token, it sends "401 - Unauthorized" response.
For missing or invalid Authorization header, it sends "400 - Bad Request".
*/
