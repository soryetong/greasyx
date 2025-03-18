package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/helper"
	"github.com/spf13/viper"
)

var secretKey []byte

func getSecretKey() []byte {
	if len(secretKey) == 0 {
		secretKey = loadSecretKey()
	}

	return secretKey
}

func loadSecretKey() []byte {
	key := viper.GetString("Jwt.SecretKey")
	if key != "" {
		return []byte(key)
	}

	randomBytes := []byte("1234567890")
	console.Echo.Warnf("⚠️ 警告: Jwt.SecretKey 为空，使用固定密钥: %s\n", string(randomBytes))

	return randomBytes
}

func GenerateJwtToken(claimsMap jwt.MapClaims) (string, error) {
	if claimsMap == nil {
		claimsMap = make(jwt.MapClaims)
	}
	claimsMap["iat"] = time.Now().Unix()
	claimsMap["nbf"] = time.Now().Unix()
	if _, ok := claimsMap["exp"]; ok == false {
		claimsMap["exp"] = time.Now().Add(time.Hour * 2).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsMap)
	tokenString, err := token.SignedString(getSecretKey())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseJwtToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getSecretKey(), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的 Token")
}

func GetTokenData[T helper.MapSupportedTypes](ctx context.Context, key string) T {
	claimsMap, ok := ctx.Value("claims").(map[string]interface{})
	if !ok {
		var zero T
		return zero
	}

	return helper.GetMapSpecificValue[T](claimsMap, key)
}
