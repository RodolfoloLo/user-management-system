package controller

import (
	"errors"
	"strconv"
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
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
	}

	if err := model.CreateUser(user); err != nil {
		// 根据具体错误类型返回不同的状态码和消息
		if errors.Is(err, model.ErrUsernameConflict) {
			return utils.Error(c, 409, "用户名已被占用")
		}
		if errors.Is(err, model.ErrEmailConflict) {
			return utils.Error(c, 409, "邮箱已被注册")
		}
		return utils.Error(c, 500, "注册失败，请稍后重试")
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

	// 2. 查询数据库
	user, err := model.GetUserByUsername(req.Username)
	if err != nil {
		// 用户不存在或数据库错误
		if errors.Is(err, model.ErrUserNotFound) {
			return utils.Error(c, 404, "用户不存在")
		}
		return utils.Error(c, 500, "系统错误，请稍后重试")
	}

	// 3. 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return utils.Error(c, 401, "密码错误")
	}

	// 4. 生成 JWT
	token, _ := utils.GenerateToken(user.ID, user.IsAdmin)
	return utils.Success(c, map[string]string{"token": token})
}

func GetUserInfo(c echo.Context) error {
	// 从 JWT 中获取用户信息
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*utils.JWTClaims)

	// 从数据库查询完整的用户信息
	user, err := model.GetUserByID(claims.UserID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return utils.Error(c, 404, "用户不存在")
		}
		return utils.Error(c, 500, "系统错误，请稍后重试")
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

	// 将需要更新的字段包装成 Map
	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}

	// 如果没有提供任何更新字段
	if len(updates) == 0 {
		return utils.Error(c, 400, "未提供任何需要更新的内容")
	}

	// 执行更新
	err := model.UpdateUser(claims.UserID, updates)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return utils.Error(c, 404, "用户不存在")
		}
		if errors.Is(err, model.ErrUsernameConflict) {
			return utils.Error(c, 409, "用户名已被占用")
		}
		if errors.Is(err, model.ErrEmailConflict) {
			return utils.Error(c, 409, "邮箱已被注册")
		}
		return utils.Error(c, 500, "更新失败，请稍后重试")
	}

	return utils.Success(c, "更新成功")
}

func AdminDeleteUser(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*utils.JWTClaims)

	if !claims.IsAdmin {
		return utils.Error(c, 403, "权限不足，仅管理员可删除用户")
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return utils.Error(c, 400, "无效的用户 ID")
	}

	deleteErr := model.DeleteUserByID(uint(id))
	if deleteErr != nil {
		if errors.Is(deleteErr, model.ErrUserNotFound) {
			return utils.Error(c, 404, "用户不存在")
		}
		return utils.Error(c, 500, "删除失败，请稍后重试")
	}

	return utils.Success(c, "删除成功")
}
