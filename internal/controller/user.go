package controller

import (
	"ums/internal/controller/params"
	"ums/internal/model"
	"ums/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Register(c echo.Context) error {
	// 1. 解析请求参数
	req := new(params.RegisterReq)
	if err := c.Bind(req); err != nil {
		return utils.Error(c, 400, "参数解析失败")
	}

	// 2. 密码加密
	hashedPassword, _ := utils.HashPassword(req.Password)
	// 3. 存入数据库
	user := model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
	}
	if err := model.DB.Create(&user).Error; err != nil {
		return utils.Error(c, 500, "注册失败")
	}
	// 4. 返回成功响应
	return utils.Success(c, "注册成功")
}

func Login(c echo.Context) error {
	// 1. 解析请求参数
	req := new(params.LoginReq)
	if err := c.Bind(req); err != nil {
		return utils.Error(c, 400, "参数解析失败")
	}

	// 2.查询数据库&验证密码
	var user model.User
	if err := model.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return utils.Error(c, 404, "用户不存在")
	}
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return utils.Error(c, 401, "密码错误")
	}
	// 3.生成 JWT
	token, _ := utils.GenerateToken(user.ID, user.IsAdmin)
	return utils.Success(c, map[string]string{"token": token})
}

func GetUserInfo(c echo.Context) error {
	// echojwt 中间件验证成功后，会自动把解析好的 token 放在 c.Get("user") 里
	// 神秘保安把Token解码成功后,会主动悄悄地调用一句： c.Set("user", 解码后的Token对象)
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*utils.JWTClaims) // 断言为我们刚才写的 JWTClaims

	var user model.User
	if err := model.DB.First(&user, claims.UserID).Error; err != nil {
		return utils.Error(c, 404, "拿不到用户信息")
	}
	return utils.Success(c, user)
}

func UpdateUser(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*utils.JWTClaims)

	req := new(params.UpdateUserReq)
	if err := c.Bind(req); err != nil {
		return utils.Error(c, 400, "参数错误")
	}

	model.DB.Model(&model.User{}).Where("id = ?", claims.UserID).Update("email", req.Email)
	return utils.Success(c, "更新成功")
}

func AdminDeleteUser(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*utils.JWTClaims)

	// 身份鉴定！
	if !claims.IsAdmin {
		return utils.Error(c, 403, "你不是管理员，没有权限删除！")
	}

	idParam := c.Param("id") // 从 URL PATH 获取被删除用户的 ID (如 /api/v1/users/3)
	// GORM 默认是软删除，也就是打个标记 DeletedAt
	if err := model.DB.Delete(&model.User{}, idParam).Error; err != nil {
		return utils.Error(c, 500, "删除失败")
	}
	return utils.Success(c, "删除成功")
}
