package utils

import (
	"errors"
	"time"

	"ums/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 定义 JWT 的载荷结构
type JWTClaims struct {
	UserID               uint `json:"user_id"`
	IsAdmin              bool `json:"is_admin"`
	jwt.RegisteredClaims      // 内嵌标准的 RegisteredClaims，包含 exp、iat 等字段
}

func GenerateToken(userID uint, isAdmin bool) (string, error) {
	claims := JWTClaims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Conf.Jwt.Expire) * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Conf.Jwt.Secret))
}

func ParseToken(tokenString string) (*JWTClaims, error) {
	// 第一步：调用框架的 ParseWithClaims (带结构体解析) 方法
	// 参数1: tokenString 就是前端传过来的那一长串火星文密文。
	// 参数2: &JWTClaims{} 相当于我们提供了一个“空盒子”。我们告诉框架：等下解密完，请把里面的数据装进这个盒子里。
	// 参数3: 这是一个匿名函数（回调函数）。框架会先拆开 Token 的最外层包装，然后转头问我们：“我要开始验证真伪了，请把你家服务器的密码拿来。”
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 在这里，我们把 config.yaml 里读取出来的全局密钥秘密地交给框架去验证
		return []byte(config.Conf.Jwt.Secret), nil
	})

	if err != nil {
		return nil, err // 解压失败，或者密码不对，直接报错退回
	}

	// 第二步：类型断言和最终确认
	// 框架解压完放入空盒子里后，由于 Go 语言的安全机制，此时盒子被贴上了一个叫 interface{}（也就是未知类型）的封条。
	// 万一黑客发来的是个别的什么奇怪玩意呢？
	// 所以我们需要用 `.(*JWTClaims)` 这一步，强行检查并撕掉封条：“我断言里面装的绝对就是 JWTClaims 这个结构的数据！”
	// 如果断言成功，ok 就会是 true。并且框架自动校验没过期的话 token.Valid 也会是 true。
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil // 一切完美，把里面装着用户ID的载荷退还给业务层
	}

	return nil, errors.New("无效的Token")
}

// 由于使用了echo-jwt中间件,这个ParseToken函数貌似大概是用不上了.
