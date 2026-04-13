package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MyClaims 自定义声明结构体并内嵌 jwt.RegisteredClaims
type MyClaims[T any] struct {
	PayLoad *T `json:"payLoad"`
	jwt.RegisteredClaims
}

var (
	ErrTokenExpired = errors.New("token已过期")
	ErrInvalidToken = errors.New("无效的token")
	// ErrTokenNotValidYet     = errors.New("token尚未生效")
	// ErrTokenMalformed       = errors.New("token格式错误")
	// ErrTokenInvalidIssuer   = errors.New("token颁发者无效")
	// ErrTokenInvalidAudience = errors.New("token接收者无效")
)

func GenToken[T any](payload *T, secret string, duration time.Duration) (string, error) {
	claims := &MyClaims[T]{
		PayLoad: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Lirous-Go-Scaffold",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken[T any](tokenString string, secret string) (*MyClaims[T], error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims[T]{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	var claims *MyClaims[T]
	if token != nil {
		claims, _ = token.Claims.(*MyClaims[T])
	}

	if err != nil {
		// 对于 JWT v5，直接判断错误类型
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			// 即使过期也返回 claims 以满足需要获取过期信息的需求
			return claims, ErrTokenExpired
		default:
			return claims, ErrInvalidToken
		}
	}

	// 验证成功
	if claims, ok := token.Claims.(*MyClaims[T]); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
