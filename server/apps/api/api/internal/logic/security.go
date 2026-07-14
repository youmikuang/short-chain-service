package logic

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// hashPassword 使用 bcrypt 对密码加盐哈希
func hashPassword(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(b), err
}

// checkPassword 校验明文密码与哈希
func checkPassword(hashed, pwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pwd))
}

// issueToken 签发 HS256 JWT（载荷含 uid）
func issueToken(secret string, expireSeconds int64, userID int64) (string, error) {
	claims := jwt.MapClaims{
		"uid": userID,
		"exp": time.Now().Add(time.Duration(expireSeconds) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
