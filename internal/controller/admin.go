package controller

import (
	"crypto/subtle" // 用于安全比较字符串，防止时序攻击
	"errors"
	"strconv"
	"ums/internal/config"
	"ums/internal/controller/params"
	"ums/internal/model"
	"ums/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func AdminBootstrap(c echo.Context) error {
	// 1. 检查初始化密钥是否配置
	if config.Conf == nil || config.Conf.Admin.BootstrapSecret == "" {
		return utils.Error(c, 503, "管理员初始化未配置")
	}

	// 2. 校验请求头中的初始化密钥
	requestSecret := c.Request().Header.Get("X-Admin-Bootstrap-Secret")
	if requestSecret == "" {
		return utils.Error(c, 401, "缺少初始化密钥")
	}

	if subtle.ConstantTimeCompare([]byte(requestSecret), []byte(config.Conf.Admin.BootstrapSecret)) != 1 {
		return utils.Error(c, 403, "初始化密钥错误")
	}

	// 3. 确认数据库里还没有管理员
	hasAdmin, err := model.HasAnyAdmin()
	if err != nil {
		return utils.Error(c, 500, "检查管理员状态失败")
	}
	if hasAdmin {
		return utils.Error(c, 409, "管理员已存在，不能再次初始化")
	}

	// 4. 解析请求参数
	req := new(params.AdminBootstrapReq)
	if err := c.Bind(req); err != nil {
		return utils.Error(c, 400, "参数解析失败")
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		return utils.Error(c, 400, "用户名、密码和邮箱不能为空")
	}

	// 5. 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.Error(c, 500, "密码加密失败")
	}

	// 6. 写入数据库
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		IsAdmin:  true,
	}

	if err := model.CreateUser(user); err != nil {
		if errors.Is(err, model.ErrUsernameConflict) {
			return utils.Error(c, 409, "用户名已被占用")
		}
		if errors.Is(err, model.ErrEmailConflict) {
			return utils.Error(c, 409, "邮箱已被注册")
		}
		return utils.Error(c, 500, "管理员初始化失败")
	}

	return utils.Success(c, "管理员初始化成功")
}

func AdminRegister(c echo.Context) error {
	// 1. 校验管理员身份
	if _, err := requireAdminClaims(c); err != nil {
		return err
	}

	// 2. 解析请求参数
	req := new(params.AdminRegisterReq)
	if err := c.Bind(req); err != nil {
		return utils.Error(c, 400, "参数解析失败")
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		return utils.Error(c, 400, "用户名、密码和邮箱不能为空")
	}

	// 3. 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.Error(c, 500, "密码加密失败")
	}

	// 4. 写入数据库
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		IsAdmin:  true,
	}

	if err := model.CreateUser(user); err != nil {
		if errors.Is(err, model.ErrUsernameConflict) {
			return utils.Error(c, 409, "用户名已被占用")
		}
		if errors.Is(err, model.ErrEmailConflict) {
			return utils.Error(c, 409, "邮箱已被注册")
		}
		return utils.Error(c, 500, "创建管理员失败")
	}

	return utils.Success(c, "管理员创建成功")
}

func AdminLogin(c echo.Context) error {
	// 1. 解析请求参数
	req := new(params.AdminLoginReq)
	if err := c.Bind(req); err != nil {
		return utils.Error(c, 400, "参数解析失败")
	}

	// 2. 查询数据库
	user, err := model.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return utils.Error(c, 404, "用户不存在")
		}
		return utils.Error(c, 500, "系统错误，请稍后重试")
	}

	if !user.IsAdmin {
		return utils.Error(c, 403, "该账号不是管理员")
	}

	// 3. 校验密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return utils.Error(c, 401, "用户名或密码错误")
	}

	// 4. 生成 JWT
	token, err := utils.GenerateToken(user.ID, true)
	if err != nil {
		return utils.Error(c, 500, "生成令牌失败")
	}

	return utils.Success(c, map[string]string{"token": token})
}

func AdminGetUserInfo(c echo.Context) error {
	// 1. 校验管理员身份
	if _, err := requireAdminClaims(c); err != nil {
		return err
	}

	// 2. 读取当前登录用户或指定用户 ID
	claims, _ := adminClaimsFromContext(c)
	if claims == nil {
		return utils.Error(c, 401, "认证信息无效")
	}

	userID := claims.UserID
	if idParam := c.Param("id"); idParam != "" {
		id, err := strconv.ParseUint(idParam, 10, 32)
		if err != nil {
			return utils.Error(c, 400, "无效的用户 ID")
		}
		userID = uint(id)
	}

	// 3. 查询数据库
	user, err := model.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return utils.Error(c, 404, "用户不存在")
		}
		return utils.Error(c, 500, "系统错误，请稍后重试")
	}

	return utils.Success(c, user)
}

func AdminCreateUser(c echo.Context) error {
	// 1. 校验管理员身份
	if _, err := requireAdminClaims(c); err != nil {
		return err
	}

	// 2. 解析请求参数
	req := new(params.AdminCreateUserReq)
	if err := c.Bind(req); err != nil {
		return utils.Error(c, 400, "参数解析失败")
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		return utils.Error(c, 400, "用户名、密码和邮箱不能为空")
	}

	// 3. 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.Error(c, 500, "密码加密失败")
	}

	// 4. 写入数据库
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		IsAdmin:  req.IsAdmin,
	}

	if err := model.CreateUser(user); err != nil {
		if errors.Is(err, model.ErrUsernameConflict) {
			return utils.Error(c, 409, "用户名已被占用")
		}
		if errors.Is(err, model.ErrEmailConflict) {
			return utils.Error(c, 409, "邮箱已被注册")
		}
		return utils.Error(c, 500, "创建用户失败")
	}

	return utils.Success(c, "创建用户成功")
}

func AdminDeleteUser(c echo.Context) error {
	// 1. 校验管理员身份
	claims, err := requireAdminClaims(c)
	if err != nil {
		return err
	}

	// 2. 解析用户 ID
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return utils.Error(c, 400, "无效的用户 ID")
	}

	if uint(id) == claims.UserID {
		return utils.Error(c, 403, "不能删除当前登录的管理员账号")
	}

	// 3. 删除用户
	if err := model.DeleteUserByID(uint(id)); err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return utils.Error(c, 404, "用户不存在")
		}
		return utils.Error(c, 500, "删除失败")
	}

	return utils.Success(c, "删除成功")
}

func AdminUpdateUser(c echo.Context) error {
	// 1. 校验管理员身份
	claims, err := requireAdminClaims(c)
	if err != nil {
		return err
	}

	// 2. 解析用户 ID
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return utils.Error(c, 400, "无效的用户 ID")
	}

	// 3. 解析请求参数
	req := new(params.AdminUpdateUserReq)
	if err := c.Bind(req); err != nil {
		return utils.Error(c, 400, "参数错误")
	}

	// 4. 收集要更新的字段
	updates := make(map[string]interface{})
	if req.Username != nil && *req.Username != "" {
		updates["username"] = *req.Username
	}
	if req.Email != nil && *req.Email != "" {
		updates["email"] = *req.Email
	}
	if req.Password != nil && *req.Password != "" {
		hashedPassword, hashErr := utils.HashPassword(*req.Password)
		if hashErr != nil {
			return utils.Error(c, 500, "密码加密失败")
		}
		updates["password"] = hashedPassword
	}
	if req.IsAdmin != nil {
		if uint(id) == claims.UserID && !*req.IsAdmin {
			return utils.Error(c, 403, "不能撤销当前登录管理员的权限")
		}
		updates["is_admin"] = *req.IsAdmin
	}

	// 5. 如果没有更新内容，直接返回
	if len(updates) == 0 {
		return utils.Error(c, 400, "未提供任何更新内容")
	}

	// 6. 执行更新
	if err := model.UpdateUser(uint(id), updates); err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return utils.Error(c, 404, "用户不存在")
		}
		if errors.Is(err, model.ErrUsernameConflict) {
			return utils.Error(c, 409, "用户名已被占用")
		}
		if errors.Is(err, model.ErrEmailConflict) {
			return utils.Error(c, 409, "邮箱已被注册")
		}
		return utils.Error(c, 500, "更新失败")
	}

	return utils.Success(c, "更新成功")
}

func adminClaimsFromContext(c echo.Context) (*utils.JWTClaims, bool) {
	// 从上下文里取出 JWT Claims
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return nil, false
	}

	claims, ok := token.Claims.(*utils.JWTClaims)
	if !ok || claims == nil {
		return nil, false
	}

	return claims, true
}

func requireAdminClaims(c echo.Context) (*utils.JWTClaims, error) {
	// 先拿到 JWT 里的管理员信息
	claims, ok := adminClaimsFromContext(c)
	if !ok {
		return nil, utils.Error(c, 401, "认证信息无效")
	}

	// 再确认是不是管理员
	if !claims.IsAdmin {
		return nil, utils.Error(c, 403, "仅管理员可访问")
	}

	return claims, nil
}
