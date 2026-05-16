package params

// 注册与创建用户
type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

// 登录参数
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 更新用户信息参数
type UpdateUserReq struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}
